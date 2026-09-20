package ukclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ukapp/internal/domain"
	"ukapp/internal/service"
)

func fakeUkServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	auth := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer tok" {
				w.WriteHeader(401)
				_, _ = w.Write([]byte(`{"code":"unauthorized","message":"no"}`))
				return
			}
			h(w, r)
		}
	}
	mux.HandleFunc("GET /houses/{fias}", auth(func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("fias") != "f-1" {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"code":"not_found","message":"no house"}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"h-1","address":"Казань, Баумана 10","organization":{"id":"uk-1","name":"УК Наш Дом"}}`))
	}))
	mux.HandleFunc("POST /incidents", auth(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		code := 201
		if body["externalRef"] == "dup" {
			code = 200
		}
		w.WriteHeader(code)
		_, _ = w.Write([]byte(`{"id":"INC-001","externalRef":"` + body["externalRef"].(string) + `","status":"accepted","updatedAt":"2026-09-19T10:00:00Z"}`))
	}))
	mux.HandleFunc("GET /incidents", auth(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("updatedSince") == "" {
			w.WriteHeader(400)
			return
		}
		_, _ = w.Write([]byte(`{"items":[{"id":"INC-001","externalRef":"7","status":"in_progress","updatedAt":"2026-09-19T10:05:00Z"}]}`))
	}))
	mux.HandleFunc("PATCH /incidents/{id}", auth(func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "INC-001" {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"code":"not_found","message":"no"}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"INC-001","externalRef":"7","status":"done","updatedAt":"2026-09-19T10:06:00Z"}`))
	}))
	// gen/uk only parses JSON200/JSON201/... when the response Content-Type contains
	// "json" (see ParseFindHouseResponse etc.) — httptest's default sniffed type is
	// text/plain, so set it explicitly here rather than in every handler.
	withJSON := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mux.ServeHTTP(w, r)
	})
	srv := httptest.NewServer(withJSON)
	t.Cleanup(srv.Close)
	return srv
}

func TestFindHouse(t *testing.T) {
	c := New(fakeUkServer(t).URL, "tok")
	h, err := c.FindHouse(context.Background(), "f-1")
	if err != nil || h.ID != "h-1" || h.OrgID != "uk-1" || h.OrgName != "УК Наш Дом" {
		t.Fatalf("got %+v err %v", h, err)
	}
	_, err = c.FindHouse(context.Background(), "nope")
	if !errors.Is(err, service.ErrUkHouseNotFound) {
		t.Fatalf("want ErrUkHouseNotFound, got %v", err)
	}
}

// TestFindHouse_200WithoutJSONBody guards against a real UK system (misconfigured proxy,
// stripped headers, ...) returning 2xx with a Content-Type gen/uk doesn't parse as JSON:
// resp.JSON200 stays nil there, and the adapter must return an error instead of panicking.
func TestFindHouse_200WithoutJSONBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /houses/{fias}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"h-1","address":"x","organization":{"id":"uk-1","name":"y"}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok")
	_, err := c.FindHouse(context.Background(), "f-1")
	if err == nil {
		t.Fatal("want error on 200 without JSON content-type, got nil")
	}
}

func TestFindHouse_Unauthorized(t *testing.T) {
	c := New(fakeUkServer(t).URL, "wrong")
	if _, err := c.FindHouse(context.Background(), "f-1"); err == nil {
		t.Fatal("want error on 401")
	}
}

func TestRegisterIncident(t *testing.T) {
	c := New(fakeUkServer(t).URL, "tok")
	id, st, err := c.RegisterIncident(context.Background(), service.UkIncident{ExternalRef: "7", HouseID: "h-1", Title: "t", Description: "d", Severity: "critical"})
	if err != nil || id != "INC-001" || st != domain.IncidentAccepted {
		t.Fatalf("got %s %s %v", id, st, err)
	}
	id, _, err = c.RegisterIncident(context.Background(), service.UkIncident{ExternalRef: "dup", HouseID: "h-1", Title: "t", Description: "d", Severity: "critical"})
	if err != nil || id != "INC-001" {
		t.Fatalf("dup: got %s %v", id, err)
	}
}

func TestIncidentUpdates(t *testing.T) {
	c := New(fakeUkServer(t).URL, "tok")
	ups, err := c.IncidentUpdates(context.Background(), time.Date(2026, 9, 19, 10, 0, 0, 0, time.UTC))
	if err != nil || len(ups) != 1 || ups[0].ExternalRef != "7" || ups[0].Status != domain.IncidentInProgress {
		t.Fatalf("got %+v err %v", ups, err)
	}
}

func TestSetStatus(t *testing.T) {
	c := New(fakeUkServer(t).URL, "tok")
	if err := c.SetStatus(context.Background(), "INC-001", domain.IncidentDone); err != nil {
		t.Fatal(err)
	}
	if err := c.SetStatus(context.Background(), "INC-999", domain.IncidentDone); err == nil {
		t.Fatal("want error on 404")
	}
}
