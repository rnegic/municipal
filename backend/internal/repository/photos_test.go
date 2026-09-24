package repository

import (
	"context"
	"testing"
	"time"
)

func TestDeleteStalePhotos(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	_, userID, incidentID := seedIncident(t, s)
	fresh, err := s.InsertPhoto(ctx, nil, userID, "image/png", []byte{1})
	if err != nil {
		t.Fatal(err)
	}
	stale, _ := s.InsertPhoto(ctx, nil, userID, "image/png", []byte{1})
	bound, _ := s.InsertPhoto(ctx, &incidentID, userID, "image/png", []byte{1})
	if _, err := s.db.ExecContext(ctx, `UPDATE incident_photo SET created_at = now() - interval '25 hours' WHERE id = ANY($1)`, []int64{stale, bound}); err != nil {
		t.Fatal(err)
	}
	n, err := s.DeleteStalePhotos(ctx, 24*time.Hour)
	if err != nil || n != 1 {
		t.Fatalf("deleted %d, %v", n, err)
	}
	for _, id := range []int64{fresh, bound} {
		if _, err := s.GetPhoto(ctx, id); err != nil {
			t.Fatalf("photo %d must stay: %v", id, err)
		}
	}
	if _, err := s.GetPhoto(ctx, stale); err != ErrNotFound {
		t.Fatalf("stale photo must be gone: %v", err)
	}
}

func TestLegacyIncidentHasNoRoutingSource(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	houseID, userID, _ := seedIncident(t, s)
	var src *string
	err := s.db.QueryRowContext(ctx, `INSERT INTO incident (house_id, title, severity, reporter_id, description)
		VALUES ($1, 'Нет воды', 'critical', $2, 'до роутинга') RETURNING routing_source`, houseID, userID).Scan(&src)
	if err != nil {
		t.Fatal(err)
	}
	if src != nil {
		t.Fatalf("legacy incident must not claim a routing source, got %q", *src)
	}
}
