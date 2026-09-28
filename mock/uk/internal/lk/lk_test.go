package lk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"mockuk/internal/store"
)

func testLK(t *testing.T) (http.Handler, *store.Store) {
	t.Helper()
	dsn := os.Getenv("MOCKUK_DATABASE_URL")
	if dsn == "" {
		t.Skip("MOCKUK_DATABASE_URL not set")
	}
	st, err := store.Open(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB().ExecContext(context.Background(), `TRUNCATE incident; ALTER SEQUENCE incident_seq RESTART`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(st.Close)
	r := gin.New()
	Mount(r, st)
	return r, st
}

func form(t *testing.T, h http.Handler, path string, v url.Values, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest("POST", path, strings.NewReader(v.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookie != "" {
		r.Header.Set("Cookie", cookie)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestLoginAndChangeStatus(t *testing.T) {
	h, st := testLK(t)
	if _, _, err := st.CreateIncident(context.Background(), store.Incident{ExternalRef: "1", HouseID: "h-1", Title: "Нет воды", Description: "x", Severity: "critical"}); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/lk", nil))
	if w.Code != 302 || w.Header().Get("Location") != "/lk/login" {
		t.Fatalf("want redirect to login, got %d %s", w.Code, w.Header().Get("Location"))
	}

	if w := form(t, h, "/lk/login", url.Values{"login": {"dispatcher"}, "password": {"wrong"}}, ""); w.Code != 401 {
		t.Fatalf("wrong password: want 401 got %d", w.Code)
	}
	w = form(t, h, "/lk/login", url.Values{"login": {"dispatcher"}, "password": {"demo1234"}}, "")
	if w.Code != 302 || w.Header().Get("Set-Cookie") == "" {
		t.Fatalf("login: %d %v", w.Code, w.Header())
	}
	cookie := strings.Split(w.Header().Get("Set-Cookie"), ";")[0]

	w = httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/lk", nil)
	r.Header.Set("Cookie", cookie)
	h.ServeHTTP(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Нет воды") || !strings.Contains(w.Body.String(), "INC-001") || !strings.Contains(w.Body.String(), "Принять") {
		t.Fatalf("list: %d %s", w.Code, w.Body)
	}

	if w := form(t, h, "/lk/incidents/INC-001/status", url.Values{"status": {"in_progress"}}, cookie); w.Code != 302 {
		t.Fatalf("status: %d %s", w.Code, w.Body)
	}
	inc, err := st.SetStatus(context.Background(), "INC-001", "in_progress")
	if err != nil || inc.Status != "in_progress" {
		t.Fatalf("status not applied: %+v %v", inc, err)
	}
}
