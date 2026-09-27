package repository

import (
	"context"
	"testing"
	"time"
)

func seedIncidentRow(t *testing.T, s *Store) int64 {
	t.Helper()
	var id int64
	err := s.db.QueryRowContext(context.Background(), `
		WITH org AS (INSERT INTO uk (external_id, name) VALUES ('uk-t', 'УК Т') RETURNING id),
		     h AS (INSERT INTO house (address_raw, house_fias_id, uk_id) SELECT 'Адрес', 'fias-t', id FROM org RETURNING id),
		     u AS (INSERT INTO app_user (max_user_id, full_name) VALUES (101, 'Житель') RETURNING id)
		INSERT INTO incident (house_id, title, severity, reporter_id, description, status)
		SELECT h.id, 'Нет воды', 'critical', u.id, 'd', 'pending' FROM h, u RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestIncidentUpdateBumpsChangeSeq(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	id := seedIncidentRow(t, s)

	var seq1, seq2 int64
	var upd1, upd2 time.Time
	if err := s.db.QueryRowContext(ctx, `SELECT change_seq, updated_at FROM incident WHERE id=$1`, id).Scan(&seq1, &upd1); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE incident SET status='accepted' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT change_seq, updated_at FROM incident WHERE id=$1`, id).Scan(&seq2, &upd2); err != nil {
		t.Fatal(err)
	}
	if seq2 <= seq1 || upd2.Before(upd1) {
		t.Fatalf("update must bump change_seq and updated_at: %d→%d %v→%v", seq1, seq2, upd1, upd2)
	}

	past := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	if _, err := s.db.ExecContext(ctx, `UPDATE incident SET updated_at=$2 WHERE id=$1`, id, past); err != nil {
		t.Fatal(err)
	}
	var got time.Time
	if err := s.db.QueryRowContext(ctx, `SELECT updated_at FROM incident WHERE id=$1`, id).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if !got.Equal(past) {
		t.Fatalf("explicit updated_at must be kept: want %v got %v", past, got)
	}
}

func TestIncidentChanges(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	id := seedIncidentRow(t, s)
	var ukID int64
	if err := s.db.QueryRowContext(ctx, `SELECT h.uk_id FROM incident i JOIN house h ON h.id=i.house_id WHERE i.id=$1`, id).Scan(&ukID); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Minute)

	if rows, err := s.IncidentChanges(ctx, ukID, 0, 100, time.Now().Add(-time.Minute)); err != nil || len(rows) != 0 {
		t.Fatalf("fresh rows must be hidden by lag: %v %d", err, len(rows))
	}
	rows, err := s.IncidentChanges(ctx, ukID, 0, 100, future)
	if err != nil || len(rows) != 1 || rows[0].ID != id || rows[0].Status != "pending" {
		t.Fatalf("first page: %v %+v", err, rows)
	}
	cursor := rows[0].ChangeSeq
	if again, _ := s.IncidentChanges(ctx, ukID, 0, 100, future); len(again) != 1 || again[0].ChangeSeq != cursor {
		t.Fatalf("same cursor must return same rows: %+v", again)
	}
	if rows, _ := s.IncidentChanges(ctx, ukID, cursor, 100, future); len(rows) != 0 {
		t.Fatalf("nothing after cursor: %+v", rows)
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE incident SET status='accepted' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	rows, _ = s.IncidentChanges(ctx, ukID, cursor, 100, future)
	if len(rows) != 1 || rows[0].Status != "accepted" || rows[0].ChangeSeq <= cursor {
		t.Fatalf("status change must appear after cursor: %+v", rows)
	}
	if rows, _ := s.IncidentChanges(ctx, ukID+1000, 0, 100, future); len(rows) != 0 {
		t.Fatalf("foreign uk must see nothing: %+v", rows)
	}
}
