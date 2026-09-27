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

func seedSecondIncident(t *testing.T, s *Store, sibling int64) int64 {
	t.Helper()
	var id int64
	err := s.db.QueryRowContext(context.Background(), `
		INSERT INTO incident (house_id, title, severity, reporter_id, description, status)
		SELECT house_id, 'Лифт', 'warning', reporter_id, 'd', 'pending' FROM incident WHERE id = $1
		RETURNING id`, sibling).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func incidentUk(t *testing.T, s *Store, id int64) int64 {
	t.Helper()
	var ukID int64
	if err := s.db.QueryRowContext(context.Background(),
		`SELECT h.uk_id FROM incident i JOIN house h ON h.id=i.house_id WHERE i.id=$1`, id).Scan(&ukID); err != nil {
		t.Fatal(err)
	}
	return ukID
}

func changeState(t *testing.T, s *Store, id int64) (*int64, time.Time) {
	t.Helper()
	var seq *int64
	var upd time.Time
	if err := s.db.QueryRowContext(context.Background(),
		`SELECT change_seq, updated_at FROM incident WHERE id=$1`, id).Scan(&seq, &upd); err != nil {
		t.Fatal(err)
	}
	return seq, upd
}

func TestIncidentChangeStamping(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	id := seedIncidentRow(t, s)

	if seq, _ := changeState(t, s, id); seq != nil {
		t.Fatalf("new incident must be unstamped, got %d", *seq)
	}
	if err := s.StampIncidentChanges(ctx); err != nil {
		t.Fatal(err)
	}
	seq1, upd1 := changeState(t, s, id)
	if seq1 == nil {
		t.Fatal("stamp must assign change_seq")
	}
	if err := s.StampIncidentChanges(ctx); err != nil {
		t.Fatal(err)
	}
	if again, updAgain := changeState(t, s, id); *again != *seq1 || !updAgain.Equal(upd1) {
		t.Fatalf("stamp must not touch stamped rows or updated_at: %d→%d %v→%v", *seq1, *again, upd1, updAgain)
	}

	if _, err := s.db.ExecContext(ctx, `UPDATE incident SET status='accepted' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	seq2, upd2 := changeState(t, s, id)
	if seq2 != nil || upd2.Before(upd1) {
		t.Fatalf("business update must unstamp and bump updated_at: %v %v→%v", seq2, upd1, upd2)
	}
	if err := s.StampIncidentChanges(ctx); err != nil {
		t.Fatal(err)
	}
	if seq3, _ := changeState(t, s, id); seq3 == nil || *seq3 <= *seq1 {
		t.Fatalf("restamp must be higher: %d then %v", *seq1, seq3)
	}

	past := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	if _, err := s.db.ExecContext(ctx, `UPDATE incident SET updated_at=$2 WHERE id=$1`, id, past); err != nil {
		t.Fatal(err)
	}
	if _, got := changeState(t, s, id); !got.Equal(past) {
		t.Fatalf("explicit updated_at must be kept: want %v got %v", past, got)
	}
}

func TestIncidentChanges(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	id := seedIncidentRow(t, s)
	ukID := incidentUk(t, s, id)

	if rows, err := s.IncidentChanges(ctx, ukID, 0, 100); err != nil || len(rows) != 0 {
		t.Fatalf("unstamped rows must not be served: %v %d", err, len(rows))
	}
	if err := s.StampIncidentChanges(ctx); err != nil {
		t.Fatal(err)
	}
	rows, err := s.IncidentChanges(ctx, ukID, 0, 100)
	if err != nil || len(rows) != 1 || rows[0].ID != id || rows[0].Status != "pending" {
		t.Fatalf("first page: %v %+v", err, rows)
	}
	cursor := rows[0].ChangeSeq
	if again, _ := s.IncidentChanges(ctx, ukID, 0, 100); len(again) != 1 || again[0].ChangeSeq != cursor {
		t.Fatalf("same cursor must return same rows: %+v", again)
	}
	if rows, _ := s.IncidentChanges(ctx, ukID, cursor, 100); len(rows) != 0 {
		t.Fatalf("nothing after cursor: %+v", rows)
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE incident SET status='accepted' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	if err := s.StampIncidentChanges(ctx); err != nil {
		t.Fatal(err)
	}
	rows, _ = s.IncidentChanges(ctx, ukID, cursor, 100)
	if len(rows) != 1 || rows[0].Status != "accepted" || rows[0].ChangeSeq <= cursor {
		t.Fatalf("status change must appear after cursor: %+v", rows)
	}
	if rows, _ := s.IncidentChanges(ctx, ukID+1000, 0, 100); len(rows) != 0 {
		t.Fatalf("foreign uk must see nothing: %+v", rows)
	}
}

func TestIncidentChanges_InFlightTransactionNotLost(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	a := seedIncidentRow(t, s)
	b := seedSecondIncident(t, s, a)
	ukID := incidentUk(t, s, a)
	if err := s.StampIncidentChanges(ctx); err != nil {
		t.Fatal(err)
	}
	start, _ := s.IncidentChanges(ctx, ukID, 0, 100)
	cursor := start[len(start)-1].ChangeSeq

	slow, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = slow.Rollback() }()
	if _, err := slow.ExecContext(ctx, `UPDATE incident SET status='accepted' WHERE id=$1`, a); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE incident SET status='accepted' WHERE id=$1`, b); err != nil {
		t.Fatal(err)
	}
	if err := s.StampIncidentChanges(ctx); err != nil {
		t.Fatal(err)
	}
	rows, _ := s.IncidentChanges(ctx, ukID, cursor, 100)
	if len(rows) != 1 || rows[0].ID != b {
		t.Fatalf("only the committed change is served: %+v", rows)
	}
	cursor = rows[0].ChangeSeq

	if err := slow.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := s.StampIncidentChanges(ctx); err != nil {
		t.Fatal(err)
	}
	rows, _ = s.IncidentChanges(ctx, ukID, cursor, 100)
	if len(rows) != 1 || rows[0].ID != a || rows[0].Status != "accepted" {
		t.Fatalf("change committed after the cursor moved must still arrive: %+v", rows)
	}
}

func unstampedCount(t *testing.T, s *Store) int {
	t.Helper()
	var n int
	if err := s.db.QueryRowContext(context.Background(), `SELECT count(*) FROM incident WHERE change_seq IS NULL`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestStampIncidentChanges_Batched(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	a := seedIncidentRow(t, s)
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO incident (house_id, title, severity, reporter_id, description, status)
		SELECT house_id, 'x', 'warning', reporter_id, 'd', 'pending' FROM incident, generate_series(1, $2)
		WHERE id = $1`, a, stampBatch); err != nil {
		t.Fatal(err)
	}
	if err := s.StampIncidentChanges(ctx); err != nil {
		t.Fatal(err)
	}
	if n := unstampedCount(t, s); n != 1 {
		t.Fatalf("one batch must stamp exactly %d rows, left %d", stampBatch, n)
	}
	if err := s.StampIncidentChanges(ctx); err != nil {
		t.Fatal(err)
	}
	if n := unstampedCount(t, s); n != 0 {
		t.Fatalf("second batch must finish, left %d", n)
	}
}

func TestStampIncidentChanges_SkipsWhenBusy(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	seedIncidentRow(t, s)
	busy, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = busy.Rollback() }()
	if _, err := busy.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, incidentStampLock); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- s.StampIncidentChanges(ctx) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("stamp must not wait for a busy stamper")
	}
	if n := unstampedCount(t, s); n != 1 {
		t.Fatalf("busy stamp must leave rows for the next poll, unstamped=%d", n)
	}
}

func TestIncidentChildWritesUnstamp(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	id := seedIncidentRow(t, s)
	var other int64
	if err := s.db.QueryRowContext(ctx, `INSERT INTO app_user (max_user_id, full_name) VALUES (102, 'Сосед') RETURNING id`).Scan(&other); err != nil {
		t.Fatal(err)
	}
	for name, write := range map[string]func() error{
		"subscribe": func() error { return s.Subscribe(ctx, id, other) },
		"confirm": func() error {
			_, err := s.UpsertConfirmation(ctx, id, other)
			return err
		},
		"photo": func() error {
			_, err := s.InsertPhoto(ctx, &id, other, "image/jpeg", []byte{0xff, 0xd8})
			return err
		},
	} {
		if err := s.StampIncidentChanges(ctx); err != nil {
			t.Fatal(err)
		}
		if err := write(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if seq, _ := changeState(t, s, id); seq != nil {
			t.Fatalf("%s must put the incident back into the feed", name)
		}
	}
}
