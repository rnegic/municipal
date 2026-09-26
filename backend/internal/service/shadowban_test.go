package service

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"ukapp/internal/domain"
)

func TestCreateIncident_SuspiciousAfterThreeFalseAlarms(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	uk := &fakeUk{}
	svc := New(s, nil, nil, uk, "testbot")
	houseID, troll := seedHouse(t, s)
	neighbour := secondUser(t, svc, 201)
	if _, err := s.DB().ExecContext(ctx, `UPDATE app_user SET house_id = $1 WHERE id = $2`, houseID, neighbour); err != nil {
		t.Fatal(err)
	}

	for i := range domain.SuspiciousAfterFalseAlarms {
		row, created, err := svc.CreateIncident(ctx, houseID, troll, NewIncident{Description: "потоп в подвале", Category: domain.CategoryWaterHeat})
		if err != nil || !created {
			t.Fatalf("#%d created=%v err=%v", i, created, err)
		}
		uk.updates = []UkIncidentUpdate{{ID: "INC-" + strconv.FormatInt(row.ID, 10), Status: domain.IncidentFalseAlarm, UpdatedAt: time.Now()}}
		if err := svc.syncStatuses(ctx); err != nil {
			t.Fatal(err)
		}
		if got := status(t, s, row.ID); got != "false_alarm" {
			t.Fatalf("#%d status=%s", i, got)
		}
	}
	for i, sus := range uk.suspicious {
		if sus {
			t.Fatalf("incident #%d sent as suspicious before the threshold", i)
		}
	}

	fake, created, err := svc.CreateIncident(ctx, houseID, troll, NewIncident{Description: "потоп в подвале", Category: domain.CategoryWaterHeat})
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	if last := uk.suspicious[len(uk.suspicious)-1]; !last {
		t.Fatal("4th incident must be registered in UK as suspicious")
	}
	if own, _ := svc.ListActiveIncidents(ctx, houseID, troll); len(own) != 1 || own[0].ID != fake.ID {
		t.Fatalf("author must still see own incident: %+v", own)
	}
	if feed, _ := svc.ListActiveIncidents(ctx, houseID, neighbour); len(feed) != 0 {
		t.Fatalf("neighbour must not see suspicious incident: %+v", feed)
	}
	genuine, created, err := svc.CreateIncident(ctx, houseID, neighbour, NewIncident{Description: "потоп в подвале", Category: domain.CategoryWaterHeat})
	if err != nil || !created || genuine.ID == fake.ID {
		t.Fatalf("neighbour report must not join suspicious incident: created=%v id=%d err=%v", created, genuine.ID, err)
	}
	if last := uk.suspicious[len(uk.suspicious)-1]; last {
		t.Fatal("neighbour incident must not be suspicious")
	}
}

func TestCreateIncident_RateLimited(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	svc := New(s, nil, nil, &fakeUk{}, "testbot")
	houseID, userID := seedHouse(t, s)
	if _, err := s.DB().ExecContext(ctx, `
		WITH inc AS (
			INSERT INTO incident (house_id, title, severity, reporter_id, description)
			SELECT $1, 't', 'warning', $2, 'd' FROM generate_series(1, $3) RETURNING id)
		INSERT INTO incident_subscription (incident_id, user_id) SELECT id, $2 FROM inc`, houseID, userID, domain.MaxReportsPerHour); err != nil {
		t.Fatal(err)
	}
	_, _, err := svc.CreateIncident(ctx, houseID, userID, NewIncident{Description: "лифт застрял", Category: domain.CategoryElevator})
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected rate limit, got %v", err)
	}
}
