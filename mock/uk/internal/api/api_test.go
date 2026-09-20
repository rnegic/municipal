package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"mockuk/internal/store"
)

func testRouter(t *testing.T) http.Handler {
	t.Helper()
	dsn := os.Getenv("MOCKUK_DATABASE_URL")
	if dsn == "" {
		t.Skip("MOCKUK_DATABASE_URL not set")
	}
	st, err := store.Open(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB().ExecContext(context.Background(), `TRUNCATE incident; ALTER SEQUENCE incident_seq RESTART;
		INSERT INTO house (id, fias_id, address, organization_id) VALUES ('h-test', 'fias-test', 'Тестовая 1', 'uk-1') ON CONFLICT (id) DO NOTHING`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(st.Close)
	return NewRouter(st, "tok", 0)
}

func do(t *testing.T, h http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestBearerRequired(t *testing.T) {
	h := testRouter(t)
	if w := do(t, h, "GET", "/houses/x", "", ""); w.Code != 401 {
		t.Fatalf("want 401, got %d", w.Code)
	}
	if w := do(t, h, "GET", "/houses/x", "", "bad"); w.Code != 401 {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestFindHouse(t *testing.T) {
	h := testRouter(t)
	w := do(t, h, "GET", "/houses/fias-test", "", "tok")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"id":"uk-1"`) || !strings.Contains(w.Body.String(), `"id":"h-test"`) {
		t.Fatalf("got %d %s", w.Code, w.Body)
	}
	if w := do(t, h, "GET", "/houses/nope", "", "tok"); w.Code != 404 {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestIncidentLifecycle(t *testing.T) {
	h := testRouter(t)
	body := `{"externalRef":"42","houseId":"h-1","title":"Нет воды","description":"x","severity":"critical"}`
	w := do(t, h, "POST", "/incidents", body, "tok")
	if w.Code != 201 {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	var inc struct{ Id, Status string }
	_ = json.Unmarshal(w.Body.Bytes(), &inc)
	if inc.Id != "INC-001" || inc.Status != "accepted" {
		t.Fatalf("bad incident %+v", inc)
	}
	// idempotent by externalRef
	if w := do(t, h, "POST", "/incidents", body, "tok"); w.Code != 200 || !strings.Contains(w.Body.String(), `"id":"INC-001"`) {
		t.Fatalf("dup: %d %s", w.Code, w.Body)
	}
	// unknown house
	if w := do(t, h, "POST", "/incidents", `{"externalRef":"43","houseId":"h-9","title":"t","description":"d","severity":"warning"}`, "tok"); w.Code != 404 {
		t.Fatalf("unknown house: want 404 got %d", w.Code)
	}
	// nothing updated after "now"
	if w := do(t, h, "GET", "/incidents?updatedSince=2999-01-01T00:00:00Z", "", "tok"); w.Code != 200 || !strings.Contains(w.Body.String(), `"items":[]`) {
		t.Fatalf("updates: %d %s", w.Code, w.Body)
	}
	// dispatcher moves to in_progress → visible in updates
	if w := do(t, h, "PATCH", "/incidents/INC-001", `{"status":"in_progress"}`, "tok"); w.Code != 200 {
		t.Fatalf("patch: %d %s", w.Code, w.Body)
	}
	w = do(t, h, "GET", "/incidents?updatedSince=2000-01-01T00:00:00Z", "", "tok")
	if !strings.Contains(w.Body.String(), `"status":"in_progress"`) || !strings.Contains(w.Body.String(), `"externalRef":"42"`) {
		t.Fatalf("updates after patch: %s", w.Body)
	}
	// done is final
	if w := do(t, h, "PATCH", "/incidents/INC-001", `{"status":"done"}`, "tok"); w.Code != 200 {
		t.Fatalf("done: %d", w.Code)
	}
	if w := do(t, h, "PATCH", "/incidents/INC-001", `{"status":"accepted"}`, "tok"); w.Code != 409 {
		t.Fatalf("reopen: want 409 got %d", w.Code)
	}
	if w := do(t, h, "PATCH", "/incidents/INC-999", `{"status":"done"}`, "tok"); w.Code != 404 {
		t.Fatalf("missing: want 404 got %d", w.Code)
	}
}

func TestFailureInjection(t *testing.T) {
	dsn := os.Getenv("MOCKUK_DATABASE_URL")
	if dsn == "" {
		t.Skip("MOCKUK_DATABASE_URL not set")
	}
	st, err := store.Open(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(st.Close)
	h := NewRouter(st, "tok", 1)
	if w := do(t, h, "GET", "/houses/x", "", "tok"); w.Code != 500 {
		t.Fatalf("failureRate=1 must always fail, got %d", w.Code)
	}
}

func TestErrorEnvelope(t *testing.T) {
	h := testRouter(t)
	// (a) invalid JSON → 400 with code:validation_failed
	w := do(t, h, "POST", "/incidents", `{not json`, "tok")
	if w.Code != 400 {
		t.Fatalf("malformed JSON: want 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"code":"validation_failed"`) {
		t.Fatalf("malformed JSON: body missing code:validation_failed, got %s", w.Body)
	}
	// (b) invalid status enum → 400 with code:validation_failed
	// First create an incident
	createBody := `{"externalRef":"e-test","houseId":"h-1","title":"Test","description":"d","severity":"critical"}`
	createW := do(t, h, "POST", "/incidents", createBody, "tok")
	if createW.Code != 201 {
		t.Fatalf("create incident: want 201, got %d", createW.Code)
	}
	var inc struct{ Id string }
	json.Unmarshal(createW.Body.Bytes(), &inc)
	// Patch with bogus status
	w = do(t, h, "PATCH", "/incidents/"+inc.Id, `{"status":"bogus"}`, "tok")
	if w.Code != 400 {
		t.Fatalf("bogus status: want 400, got %d %s", w.Code, w.Body)
	}
	if !strings.Contains(w.Body.String(), `"code":"validation_failed"`) {
		t.Fatalf("bogus status: body missing code:validation_failed, got %s", w.Body)
	}
}

func TestBearerEmptyTokenFailsClosed(t *testing.T) {
	dsn := os.Getenv("MOCKUK_DATABASE_URL")
	if dsn == "" {
		t.Skip("MOCKUK_DATABASE_URL not set")
	}
	st, err := store.Open(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(st.Close)
	// Router with empty token
	h := NewRouter(st, "", 0)
	// Request with "Authorization: Bearer " (empty token) should fail
	r := httptest.NewRequest("GET", "/houses/x", nil)
	r.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("empty bearer token: want 401, got %d", w.Code)
	}
}
