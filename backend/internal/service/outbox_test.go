package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"ukapp/internal/maxclient"
	"ukapp/internal/repository"
)

func testStore(t *testing.T) *repository.Store {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	s, err := repository.Open(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.DB().ExecContext(context.Background(),
		`TRUNCATE outbox_message, incident_confirmation, incident_subscription, incident, app_user, house, uk RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

type sentMsg struct {
	UserID string
	Body   map[string]any
}

func fakeMax(t *testing.T, failFirst int) (*maxclient.Client, *[]sentMsg) {
	t.Helper()
	var mu sync.Mutex
	var sent []sentMsg
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		if calls <= failFirst {
			w.WriteHeader(503)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		sent = append(sent, sentMsg{UserID: r.URL.Query().Get("user_id"), Body: body})
		_, _ = w.Write([]byte(`{"message":{}}`))
	}))
	t.Cleanup(srv.Close)
	c := maxclient.NewClient("tok")
	c.BaseURL = srv.URL
	return c, &sent
}

func outboxCount(t *testing.T, s *repository.Store, status string) int {
	t.Helper()
	var n int
	if err := s.DB().QueryRowContext(context.Background(), `SELECT count(*) FROM outbox_message WHERE status=$1`, status).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestOutboxWorker_DeliversAndRetries(t *testing.T) {
	repo := testStore(t)
	maxc, sent := fakeMax(t, 1) // first call fails, then succeeds
	svc := New(repo, maxc, nil, nil)
	ctx := context.Background()

	tx, err := repo.DB().BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Enqueue(ctx, tx, []int64{7, 8}, "test", repository.OutboxPayload{Text: "ping"}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if got := outboxCount(t, repo, "pending"); got != 2 {
		t.Fatalf("pending=%d want 2", got)
	}

	wctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	go svc.RunOutboxWorker(wctx)

	deadline := time.Now().Add(7 * time.Second)
	for time.Now().Before(deadline) && outboxCount(t, repo, "sent") < 2 {
		// make retry immediate for the test
		_, _ = repo.DB().ExecContext(ctx, `UPDATE outbox_message SET next_attempt_at = now() WHERE status='pending'`)
		time.Sleep(200 * time.Millisecond)
	}
	if sentCount, failed := outboxCount(t, repo, "sent"), outboxCount(t, repo, "failed"); sentCount != 2 || failed != 0 {
		t.Fatalf("sent=%d failed=%d delivered=%d", sentCount, failed, len(*sent))
	}
}
