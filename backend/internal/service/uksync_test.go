package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"ukapp/internal/domain"
	"ukapp/internal/repository"
)

type fakeUk struct {
	mu         sync.Mutex
	registered int
	updates    []UkIncidentUpdate
	updatesErr error
	setStatus  []string
	fail       bool
}

func (f *fakeUk) FindHouse(_ context.Context, fias string) (UkHouse, error) {
	return UkHouse{ID: "h-" + fias, Address: fias, OrgID: "uk-1", OrgName: "Демо УК"}, nil
}

func (f *fakeUk) RegisterIncident(_ context.Context, in UkIncident) (string, domain.IncidentStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail {
		return "", "", context.DeadlineExceeded
	}
	f.registered++
	return "INC-" + in.ExternalRef, domain.IncidentAccepted, nil
}

func (f *fakeUk) IncidentUpdates(context.Context, time.Time) ([]UkIncidentUpdate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.updatesErr != nil {
		return nil, f.updatesErr
	}
	return f.updates, nil
}

func (f *fakeUk) SetStatus(_ context.Context, id string, st domain.IncidentStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.setStatus = append(f.setStatus, id+":"+string(st))
	return nil
}

// seedResidentWithIncident: one UK, one house, one resident, one incident (via repo, no HTTP).
func seedResidentWithIncident(t *testing.T, s *repository.Store, maxID int64) (incidentID, userID int64) {
	t.Helper()
	ctx := context.Background()
	u, err := s.UpsertUser(ctx, domain.InitUser{ID: maxID, FirstName: "R"})
	if err != nil {
		t.Fatal(err)
	}
	ukID, err := s.UpsertUk(ctx, "uk-1", "Демо УК")
	if err != nil {
		t.Fatal(err)
	}
	houseID, err := s.BindHouse(ctx, u.ID, "Баумана 10", "f-10", ukID, "h-f-10")
	if err != nil {
		t.Fatal(err)
	}
	incidentID, err = s.CreateIncident(ctx, houseID, u.ID, "Нет воды", "x", domain.SeverityCritical, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return incidentID, u.ID
}

func externalID(t *testing.T, s *repository.Store, id int64) *string {
	t.Helper()
	var ext *string
	if err := s.DB().QueryRowContext(context.Background(), `SELECT external_id FROM incident WHERE id=$1`, id).Scan(&ext); err != nil {
		t.Fatal(err)
	}
	return ext
}

func TestSyncUnregistered_RegistersAndIsIdempotent(t *testing.T) {
	s := testStore(t)
	uk := &fakeUk{}
	svc := New(s, nil, nil, uk)
	incID, _ := seedResidentWithIncident(t, s, 1)
	if ext := externalID(t, s, incID); ext != nil {
		t.Fatalf("fresh incident must be unregistered, got %s", *ext)
	}
	if err := svc.syncUnregistered(context.Background()); err != nil {
		t.Fatal(err)
	}
	ext := externalID(t, s, incID)
	if ext == nil || *ext != "INC-1" {
		t.Fatalf("want INC-1, got %v", ext)
	}
	if err := svc.syncUnregistered(context.Background()); err != nil {
		t.Fatal(err)
	}
	if uk.registered != 1 {
		t.Fatalf("second sync must not re-register, calls=%d", uk.registered)
	}
}

func TestSyncUnregistered_UkDownLeavesNull(t *testing.T) {
	s := testStore(t)
	uk := &fakeUk{fail: true}
	svc := New(s, nil, nil, uk)
	incID, _ := seedResidentWithIncident(t, s, 2)
	if err := svc.syncUnregistered(context.Background()); err == nil {
		t.Fatal("want error when uk is down")
	}
	if ext := externalID(t, s, incID); ext != nil {
		t.Fatalf("must stay unregistered, got %s", *ext)
	}
}

func status(t *testing.T, s *repository.Store, id int64) string {
	t.Helper()
	var st string
	if err := s.DB().QueryRowContext(context.Background(), `SELECT status FROM incident WHERE id=$1`, id).Scan(&st); err != nil {
		t.Fatal(err)
	}
	return st
}

func TestSyncStatuses_AppliesOnceAndNotifies(t *testing.T) {
	s := testStore(t)
	uk := &fakeUk{}
	svc := New(s, nil, nil, uk)
	incID, _ := seedResidentWithIncident(t, s, 3)
	if err := svc.syncUnregistered(context.Background()); err != nil {
		t.Fatal(err)
	}
	uk.updates = []UkIncidentUpdate{{ID: "INC-1", ExternalRef: "1", Status: domain.IncidentInProgress, UpdatedAt: time.Now()}}

	if err := svc.syncStatuses(context.Background()); err != nil {
		t.Fatal(err)
	}
	if st := status(t, s, incID); st != "in_progress" {
		t.Fatalf("want in_progress, got %s", st)
	}
	if n := outboxCount(t, s, "pending"); n != 1 {
		t.Fatalf("want 1 push to the reporter, got %d", n)
	}
	// same update again → no-op, no duplicate push
	if err := svc.syncStatuses(context.Background()); err != nil {
		t.Fatal(err)
	}
	if n := outboxCount(t, s, "pending"); n != 1 {
		t.Fatalf("duplicate push: %d", n)
	}
}

func TestSyncStatuses_DoneIsFinal(t *testing.T) {
	s := testStore(t)
	uk := &fakeUk{}
	svc := New(s, nil, nil, uk)
	incID, _ := seedResidentWithIncident(t, s, 4)
	if err := svc.syncUnregistered(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().ExecContext(context.Background(), `UPDATE incident SET status='done' WHERE id=$1`, incID); err != nil {
		t.Fatal(err)
	}
	uk.updates = []UkIncidentUpdate{{ID: "INC-1", ExternalRef: "1", Status: domain.IncidentVerifying, UpdatedAt: time.Now()}}
	if err := svc.syncStatuses(context.Background()); err != nil {
		t.Fatal(err)
	}
	if st := status(t, s, incID); st != "done" {
		t.Fatalf("done must not be reopened, got %s", st)
	}
	if n := outboxCount(t, s, "pending"); n != 0 {
		t.Fatalf("no push expected, got %d", n)
	}
}

func TestConfirm_PushesDoneToUk(t *testing.T) {
	s := testStore(t)
	uk := &fakeUk{}
	svc := New(s, nil, nil, uk)
	incID, userID := seedResidentWithIncident(t, s, 5)
	if err := svc.syncUnregistered(context.Background()); err != nil {
		t.Fatal(err)
	}
	// second subscriber so MinConfirmations=2 is reachable
	u2, err := s.UpsertUser(context.Background(), domain.InitUser{ID: 6, FirstName: "S"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Subscribe(context.Background(), incID, u2.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().ExecContext(context.Background(), `UPDATE incident SET status='verifying' WHERE id=$1`, incID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ConfirmIncident(context.Background(), incID, userID); err != nil {
		t.Fatal(err)
	}
	st, _, err := svc.ConfirmIncident(context.Background(), incID, u2.ID)
	if err != nil || st != domain.IncidentDone {
		t.Fatalf("want done, got %s %v", st, err)
	}
	if len(uk.setStatus) != 1 || uk.setStatus[0] != "INC-1:done" {
		t.Fatalf("want done pushed to uk once, got %v", uk.setStatus)
	}
}

func TestSyncStatuses_CursorNotAdvancedOnError(t *testing.T) {
	s := testStore(t)
	uk := &fakeUk{}
	svc := New(s, nil, nil, uk)

	someTime := time.Now().Add(-time.Hour)
	svc.ukSince = someTime
	uk.updatesErr = errors.New("uk unreachable")
	if err := svc.syncStatuses(context.Background()); err == nil {
		t.Fatal("want error from IncidentUpdates")
	}
	if !svc.ukSince.Equal(someTime) {
		t.Fatalf("cursor must not move on error, want %v got %v", someTime, svc.ukSince)
	}

	uk.updatesErr = nil
	t2 := someTime.Add(time.Minute)
	uk.updates = []UkIncidentUpdate{{ID: "INC-nope", ExternalRef: "999", Status: domain.IncidentInProgress, UpdatedAt: t2}}
	if err := svc.syncStatuses(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !svc.ukSince.Equal(t2) {
		t.Fatalf("cursor must advance to last update on success, want %v got %v", t2, svc.ukSince)
	}
}
