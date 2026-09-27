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
		INSERT INTO incident (house_id, title, severity, reporter_id, description)
		SELECT h.id, 'Нет воды', 'critical', u.id, 'd' FROM h, u RETURNING id`).Scan(&id)
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
