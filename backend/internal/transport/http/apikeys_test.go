package http

import (
	"encoding/json"
	"net/http"
	"testing"
)

func issueKey(t *testing.T, srv http.Handler, token, name string) (id, key string) {
	t.Helper()
	w := serve(srv, bearerReq("POST", "/api/uk/api-keys", `{"name":"`+name+`"}`, token))
	if w.Code != 201 {
		t.Fatalf("issue key: %d %s", w.Code, w.Body)
	}
	var out struct{ Id, Key string }
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return out.Id, out.Key
}

func TestApiKeys_IssueUseRevoke(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	seedDispatcher(t, s, "uk-1", "1655000003", "2099-12-31", true)
	seedDispatcher(t, s, "uk-2", "7707083893", "2099-12-31", true)
	token := dispatcherToken(t, srv, "1655000003")
	otherToken := dispatcherToken(t, srv, "7707083893")

	id, key := issueKey(t, srv, token, "1С офис")
	if len(key) < 20 || key[:8] != "uk_live_" {
		t.Fatalf("key: %q", key)
	}

	var me struct{ User struct{ Role, FullName string } }
	w := serve(srv, bearerReq("GET", "/api/me", "", key))
	_ = json.Unmarshal(w.Body.Bytes(), &me)
	if w.Code != 200 || me.User.Role != "uk_dispatcher" || me.User.FullName != "1С-интеграция" {
		t.Fatalf("me under key: %d %s", w.Code, w.Body)
	}
	if w := serve(srv, bearerReq("GET", "/api/uk/queue", "", key)); w.Code != 200 {
		t.Fatalf("queue under key: %d %s", w.Code, w.Body)
	}

	for _, r := range []*http.Request{
		bearerReq("POST", "/api/uk/api-keys", `{"name":"x"}`, key),
		bearerReq("GET", "/api/uk/api-keys", "", key),
		bearerReq("DELETE", "/api/uk/api-keys/"+id, "", key),
	} {
		if w := serve(srv, r); w.Code != 403 {
			t.Fatalf("%s %s under key: want 403 got %d", r.Method, r.URL.Path, w.Code)
		}
	}

	var list struct {
		Items []map[string]any
	}
	w = serve(srv, bearerReq("GET", "/api/uk/api-keys", "", token))
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if w.Code != 200 || len(list.Items) != 1 || list.Items[0]["key"] != nil || list.Items[0]["lastUsedAt"] == nil {
		t.Fatalf("list: %d %s", w.Code, w.Body)
	}

	if w := serve(srv, bearerReq("DELETE", "/api/uk/api-keys/"+id, "", otherToken)); w.Code != 404 {
		t.Fatalf("revoke foreign: want 404 got %d", w.Code)
	}
	if w := serve(srv, bearerReq("DELETE", "/api/uk/api-keys/"+id, "", token)); w.Code != 204 {
		t.Fatalf("revoke: %d %s", w.Code, w.Body)
	}
	if w := serve(srv, bearerReq("GET", "/api/uk/queue", "", key)); w.Code != 401 {
		t.Fatalf("revoked key: want 401 got %d", w.Code)
	}

	if w := esiaLogin(t, srv, "1655000003", dispatcherPassword); w.Code != 200 {
		t.Fatalf("human dispatcher login after key issue: %d %s", w.Code, w.Body)
	}
}

func TestApiKeys_ValidationAndLimit(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	seedDispatcher(t, s, "uk-1", "1655000003", "2099-12-31", true)
	token := dispatcherToken(t, srv, "1655000003")
	if w := serve(srv, bearerReq("POST", "/api/uk/api-keys", `{"name":"  "}`, token)); w.Code != 400 {
		t.Fatalf("blank name: want 400 got %d", w.Code)
	}
	for i := 0; i < 10; i++ {
		issueKey(t, srv, token, "k")
	}
	if w := serve(srv, bearerReq("POST", "/api/uk/api-keys", `{"name":"k"}`, token)); w.Code != 422 {
		t.Fatalf("11th key: want 422 got %d", w.Code)
	}
	if w := serve(srv, bearerReq("GET", "/api/uk/queue", "", "uk_live_nope")); w.Code != 401 {
		t.Fatalf("unknown key: want 401 got %d", w.Code)
	}
}

func TestApiKeys_RateLimited(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	seedDispatcher(t, s, "uk-1", "1655000003", "2099-12-31", true)
	_, key := issueKey(t, srv, dispatcherToken(t, srv, "1655000003"), "k")
	got429 := false
	for i := 0; i < 30 && !got429; i++ {
		got429 = serve(srv, bearerReq("GET", "/api/uk/queue", "", key)).Code == 429
	}
	if !got429 {
		t.Fatal("burst of 30 requests must hit 429")
	}
}
