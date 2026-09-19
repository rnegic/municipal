package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"ukapp/internal/dadata"
	"ukapp/internal/repository"
	"ukapp/internal/service"
)

// These are integration tests exercising the whole stack (transport → service → repository)
// against a real Postgres, per CLAUDE.md testing policy. Raw SQL setup below (assigning the
// dispatcher role) has no HTTP endpoint, so it goes straight through repository.Store.DB() —
// that's why depguard exempts _test.go files from the transport→repository boundary.

func testStore(t *testing.T) *repository.Store {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	s, err := repository.Open(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.DB().ExecContext(context.Background(),
		`TRUNCATE outbox_message, event_response, event, incident_confirmation, incident_subscription, incident, app_user, house, uk RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s
}

func seedUK(t *testing.T, s *repository.Store) {
	t.Helper()
	if _, err := s.DB().ExecContext(context.Background(), `INSERT INTO uk (name) VALUES ('Демо УК')`); err != nil {
		t.Fatal(err)
	}
}

// fakeDadata echoes the query back as both the normalized address and the house_fias_id, so
// tests can use a plain string (e.g. "f-10") as a stand-in house identity, same as before.
func fakeDadata(t *testing.T) *dadata.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"suggestions":[{"value":"` + body.Query + `","data":{"house_fias_id":"` + body.Query + `"}}]}`))
	}))
	t.Cleanup(srv.Close)
	c := dadata.NewClient("k")
	c.BaseURL = srv.URL
	return c
}

func newTestServer(t *testing.T, repo *repository.Store) http.Handler {
	t.Helper()
	return NewServer(service.New(repo, nil, fakeDadata(t)), testBotToken)
}

func itoa(n int64) string              { return strconv.FormatInt(n, 10) }
func stringsReader(s string) io.Reader { return strings.NewReader(s) }

const testBotToken = "test-bot-token"

// signInitData builds a valid initData string the way MAX does (see docs/webapps/validation).
// auth_date is always "now" — ValidateInitData rejects anything older than 24h.
func signInitData(t *testing.T, pairs map[string]string) string {
	t.Helper()
	if _, ok := pairs["auth_date"]; !ok {
		pairs["auth_date"] = strconv.FormatInt(time.Now().Unix(), 10)
	}
	keys := make([]string, 0, len(pairs))
	for k := range pairs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	check := ""
	for i, k := range keys {
		if i > 0 {
			check += "\n"
		}
		check += k + "=" + pairs[k]
	}
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(testBotToken))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(check))
	hash := hex.EncodeToString(mac.Sum(nil))
	q := url.Values{}
	for k, v := range pairs {
		q.Set(k, v)
	}
	q.Set("hash", hash)
	return q.Encode()
}

func authedReq(t *testing.T, method, path, body string, maxUserID int64, name string) *http.Request {
	t.Helper()
	raw := signInitData(t, map[string]string{
		"user": `{"id":` + itoa(maxUserID) + `,"first_name":"` + name + `","last_name":"T"}`,
	})
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, stringsReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	r.Header.Set("Authorization", "tma "+raw)
	return r
}

// bindUser binds a house identified by addr (used as both DaData value and house_fias_id).
func bindUser(t *testing.T, srv http.Handler, maxID int64, addr string) string {
	t.Helper()
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/houses/bind", `{"address":"`+addr+`"}`, maxID, "U"))
	if w.Code != 200 {
		t.Fatalf("bind: %d %s", w.Code, w.Body)
	}
	var h struct{ Id string }
	_ = json.Unmarshal(w.Body.Bytes(), &h)
	return h.Id
}

func incidentStatus(t *testing.T, s *repository.Store, id int64) string {
	t.Helper()
	var st string
	if err := s.DB().QueryRowContext(context.Background(), `SELECT status FROM incident WHERE id=$1`, id).Scan(&st); err != nil {
		t.Fatal(err)
	}
	return st
}

// makeDispatcher upserts the user (first authed request) and promotes it to uk_dispatcher of the seeded UK.
func makeDispatcher(t *testing.T, srv http.Handler, s *repository.Store, maxID int64) {
	t.Helper()
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/me", "", maxID, "D"))
	if w.Code != 200 {
		t.Fatalf("me: %d %s", w.Code, w.Body)
	}
	_, err := s.DB().ExecContext(context.Background(),
		`UPDATE app_user SET role='uk_dispatcher', uk_id=(SELECT id FROM uk ORDER BY id LIMIT 1) WHERE max_user_id=$1`, maxID)
	if err != nil {
		t.Fatal(err)
	}
}

// setStatus drives PATCH /api/incidents/{id}/status as the dispatcher and returns the HTTP code.
func setStatus(t *testing.T, srv http.Handler, dispatcherMaxID int64, incID, status string) int {
	t.Helper()
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "PATCH", "/api/incidents/"+incID+"/status", `{"status":"`+status+`"}`, dispatcherMaxID, "D"))
	return w.Code
}
