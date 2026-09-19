package http

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

type meResponse struct {
	User struct {
		Id       string
		FullName string
		Role     string
	}
	House *struct {
		Id      string
		Address string
	}
}

func TestMe_CreatesOnFirstCall(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/me", "", 500, "Anna"))
	if w.Code != 200 {
		t.Fatalf("code %d body %s", w.Code, w.Body)
	}
	var me meResponse
	if err := json.Unmarshal(w.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me.User.FullName != "Anna T" || me.User.Role != "resident" || me.House != nil {
		t.Fatalf("bad me %+v", me)
	}

	// second call: same row, not a duplicate
	srv.ServeHTTP(httptest.NewRecorder(), authedReq(t, "GET", "/api/me", "", 500, "Anna"))
	var n int
	if err := s.DB().QueryRowContext(t.Context(), `SELECT count(*) FROM app_user`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 user, got %d", n)
	}
}

func TestMe_Unauthorized(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest("GET", "/api/me", nil))
	if w.Code != 401 {
		t.Fatalf("code %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"code":"unauthorized"`) {
		t.Fatalf("expected JSON error, got %s", w.Body)
	}
}

func TestBindHouse(t *testing.T) {
	s := testStore(t)
	seedUK(t, s)
	srv := newTestServer(t, s)

	h1 := bindUser(t, srv, 501, "abc-123")
	if h1 == "" {
		t.Fatal("house id not set")
	}
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/me", "", 501, "Ivan"))
	var me meResponse
	_ = json.Unmarshal(w.Body.Bytes(), &me)
	if me.House == nil || me.House.Id != h1 {
		t.Fatalf("house not bound in profile: %+v", me)
	}

	// second resident, same normalized address → same house row
	h2 := bindUser(t, srv, 502, "abc-123")
	if h2 != h1 {
		t.Fatalf("same address must map to same house: %v vs %v", h1, h2)
	}

	// empty address → 400 JSON
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/houses/bind", `{"address":""}`, 501, "Ivan"))
	if w.Code != 400 {
		t.Fatalf("code %d body %s", w.Code, w.Body)
	}
}
