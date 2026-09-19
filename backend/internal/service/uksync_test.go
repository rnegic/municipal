package service

import (
	"context"
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
