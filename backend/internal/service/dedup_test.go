package service

import (
	"context"
	"errors"
	"testing"

	"ukapp/internal/domain"
)

type fakeMatcher struct {
	id    int64
	err   error
	calls int
	req   domain.OpenIncident
	cands []domain.OpenIncident
}

func (f *fakeMatcher) Match(_ context.Context, req domain.OpenIncident, cands []domain.OpenIncident) (domain.DedupMatch, error) {
	f.calls++
	f.req, f.cands = req, cands
	if f.err != nil {
		return domain.DedupMatch{}, f.err
	}
	return domain.DedupMatch{IncidentID: f.id, P: 0.97, ModelVersion: "test-model@1"}, nil
}

func secondUser(t *testing.T, svc *Service, id int64) int64 {
	t.Helper()
	u, err := svc.repo.UpsertUser(context.Background(), domain.InitUser{ID: id, FirstName: "S"})
	if err != nil {
		t.Fatal(err)
	}
	return u.ID
}

func TestCreateIncident_ModelJoinsAcrossCategories(t *testing.T) {
	s := testStore(t)
	m := &fakeMatcher{}
	svc := New(s, nil, nil, &fakeUk{}, "testbot").WithMatcher(m)
	houseID, userID := seedHouse(t, s)
	ctx := context.Background()
	first, created, err := svc.CreateIncident(ctx, houseID, userID, NewIncident{Description: "нет горячей воды с утра", Category: domain.CategoryWaterHeat})
	if err != nil || !created || m.calls != 0 {
		t.Fatalf("first: created=%v err=%v calls=%d", created, err, m.calls)
	}
	m.id = first.ID
	row, created, err := svc.CreateIncident(ctx, houseID, secondUser(t, svc, 201), NewIncident{Description: "из крана течёт ледяная", Category: domain.CategoryElevator})
	if err != nil || created || row.ID != first.ID {
		t.Fatalf("want join into %d: id=%d created=%v err=%v", first.ID, row.ID, created, err)
	}
	if len(m.cands) != 1 || m.cands[0].ID != first.ID || m.cands[0].Description != "нет горячей воды с утра" || m.req.Description != "из крана течёт ледяная" {
		t.Fatalf("req=%+v cands=%+v", m.req, m.cands)
	}
	got := serviceReports(t, s)
	if len(got) != 2 || got[1].outcome != "joined" || got[1].dedupVersion != "test-model@1" {
		t.Fatalf("%+v", got)
	}
}

func TestCreateIncident_ModelSaysNewOverridesRule(t *testing.T) {
	s := testStore(t)
	svc := New(s, nil, nil, &fakeUk{}, "testbot").WithMatcher(&fakeMatcher{})
	houseID, userID := seedHouse(t, s)
	ctx := context.Background()
	if _, _, err := svc.CreateIncident(ctx, houseID, userID, NewIncident{Description: "нет воды с утра", Category: domain.CategoryWaterHeat}); err != nil {
		t.Fatal(err)
	}
	_, created, err := svc.CreateIncident(ctx, houseID, secondUser(t, svc, 201), NewIncident{Description: "нет воды и у нас", Category: domain.CategoryWaterHeat})
	if err != nil || !created {
		t.Fatalf("model said no duplicate: created=%v err=%v", created, err)
	}
	got := serviceReports(t, s)
	if len(got) != 2 || got[1].outcome != "created" || got[1].dedupVersion != "test-model@1" {
		t.Fatalf("%+v", got)
	}
}

func TestCreateIncident_ModelErrorFallsBackToRule(t *testing.T) {
	s := testStore(t)
	m := &fakeMatcher{err: errors.New("down")}
	svc := New(s, nil, nil, &fakeUk{}, "testbot").WithMatcher(m)
	houseID, userID := seedHouse(t, s)
	ctx := context.Background()
	first, _, err := svc.CreateIncident(ctx, houseID, userID, NewIncident{Description: "нет воды с утра", Category: domain.CategoryWaterHeat})
	if err != nil {
		t.Fatal(err)
	}
	row, created, err := svc.CreateIncident(ctx, houseID, secondUser(t, svc, 201), NewIncident{Description: "нет воды и у нас", Category: domain.CategoryWaterHeat})
	if err != nil || created || row.ID != first.ID {
		t.Fatalf("rule must join: id=%d created=%v err=%v", row.ID, created, err)
	}
	if m.calls != 1 {
		t.Fatalf("matcher must be called exactly once: calls=%d", m.calls)
	}
	got := serviceReports(t, s)
	if len(got) != 2 || got[1].dedupVersion != domain.DedupVersion {
		t.Fatalf("%+v", got)
	}
}

func TestCreateIncident_ModelSkipsAlreadySubscribedIncident(t *testing.T) {
	s := testStore(t)
	m := &fakeMatcher{}
	svc := New(s, nil, nil, &fakeUk{}, "testbot").WithMatcher(m)
	houseID, userID := seedHouse(t, s)
	ctx := context.Background()
	first, created, err := svc.CreateIncident(ctx, houseID, userID, NewIncident{Description: "нет горячей воды с утра", Category: domain.CategoryWaterHeat})
	if err != nil || !created {
		t.Fatalf("first: created=%v err=%v", created, err)
	}
	m.id = first.ID
	row, created, err := svc.CreateIncident(ctx, houseID, userID, NewIncident{Description: "сломан лифт", Category: domain.CategoryElevator})
	if err != nil || !created || row.ID == first.ID {
		t.Fatalf("must not join own subscribed incident: id=%d created=%v err=%v", row.ID, created, err)
	}
	if m.calls != 0 || len(m.cands) != 0 {
		t.Fatalf("matcher must not see already-subscribed incident: calls=%d cands=%+v", m.calls, m.cands)
	}
	got := serviceReports(t, s)
	if len(got) != 2 || got[0].outcome != "created" || got[1].outcome != "created" {
		t.Fatalf("%+v", got)
	}
}

func TestCreateIncident_ModelRetryDoesNotDuplicateReport(t *testing.T) {
	s := testStore(t)
	m := &fakeMatcher{}
	svc := New(s, nil, nil, &fakeUk{}, "testbot").WithMatcher(m)
	houseID, userID := seedHouse(t, s)
	first, _, err := svc.CreateIncident(context.Background(), houseID, userID, NewIncident{Description: "нет воды с утра", Category: domain.CategoryWaterHeat})
	if err != nil {
		t.Fatal(err)
	}
	m.id = first.ID
	if _, _, err := svc.CreateIncident(context.Background(), houseID, userID, NewIncident{Description: "нет воды с утра", Category: domain.CategoryWaterHeat}); err != nil {
		t.Fatal(err)
	}
	if got := serviceReports(t, s); len(got) != 1 {
		t.Fatalf("retry must not add a report: %+v", got)
	}
}
