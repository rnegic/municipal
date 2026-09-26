package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"ukapp/internal/dadata"
	"ukapp/internal/domain"
	"ukapp/internal/repository"
	"ukapp/internal/service"
)

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
		`TRUNCATE outbox_message, incident_confirmation, incident_subscription, incident, app_user, house, uk RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s
}

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

type fakeUk struct {
	mu         sync.Mutex
	registered []service.UkIncident
	updates    []service.UkIncidentUpdate
	setStatus  []string
}

func (f *fakeUk) FindHouse(_ context.Context, fias, _ string) (service.UkHouse, error) {
	if fias == "unknown" {
		return service.UkHouse{}, service.ErrUkHouseNotFound
	}
	phone := "+7 (843) 200-00-00"
	emergency := "+7 (843) 200-01-01"
	return service.UkHouse{
		ID: "h-" + fias, Address: fias,
		Org: service.UkOrg{ExternalID: "uk-1", Name: "Демо УК", Phone: &phone, EmergencyPhone: &emergency},
	}, nil
}

func (f *fakeUk) RegisterIncident(_ context.Context, in service.UkIncident) (string, domain.IncidentStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.registered = append(f.registered, in)
	return "INC-" + in.ExternalRef, domain.IncidentAccepted, nil
}

func (f *fakeUk) IncidentUpdates(context.Context, time.Time) ([]service.UkIncidentUpdate, error) {
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

type fakeClassifier struct {
	p   domain.Prediction
	err error
}

func (f *fakeClassifier) Classify(context.Context, string) (domain.Prediction, error) {
	return f.p, f.err
}

func newTestServer(t *testing.T, repo *repository.Store) http.Handler {
	t.Helper()
	return newTestServerWith(t, repo, &fakeClassifier{err: errors.New("model off")})
}

func newTestServerWith(t *testing.T, repo *repository.Store, cls service.Classifier) http.Handler {
	t.Helper()
	svc := service.New(repo, nil, fakeDadata(t), &fakeUk{}, "testbot").WithClassifier(cls, 0.7)
	return NewServer(svc, testBotToken)
}

func itoa(n int64) string              { return strconv.FormatInt(n, 10) }
func stringsReader(s string) io.Reader { return strings.NewReader(s) }

const testBotToken = "test-bot-token"

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

func setVerifying(t *testing.T, s *repository.Store, id int64) {
	t.Helper()
	if _, err := s.DB().ExecContext(context.Background(), `UPDATE incident SET status='verifying' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
}
