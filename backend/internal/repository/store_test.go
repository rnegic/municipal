package repository

import (
	"context"
	"os"
	"testing"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	s, err := Open(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.db.ExecContext(context.Background(),
		`TRUNCATE outbox_message, event_response, event, incident_confirmation, incident_subscription, incident, app_user, house, uk RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.db.Close() })
	return s
}

func TestOpenAppliesSchema(t *testing.T) {
	s := testStore(t)
	var n int
	if err := s.db.QueryRowContext(context.Background(), `SELECT count(*) FROM incident`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected empty incident table, got %d", n)
	}
}
