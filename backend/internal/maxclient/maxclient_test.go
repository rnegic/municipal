package maxclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_SendMessageFormat(t *testing.T) {
	var got struct {
		UserID string
		Body   map[string]any
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "tok" {
			t.Errorf("bad auth %q", r.Header.Get("Authorization"))
		}
		got.UserID = r.URL.Query().Get("user_id")
		_ = json.NewDecoder(r.Body).Decode(&got.Body)
		_, _ = w.Write([]byte(`{"message":{}}`))
	}))
	defer srv.Close()

	c := NewClient("tok")
	c.BaseURL = srv.URL
	if err := c.SendMessage(context.Background(), 42, "hi"); err != nil {
		t.Fatal(err)
	}
	if got.UserID != "42" || got.Body["text"] != "hi" {
		t.Fatalf("bad msg %+v", got)
	}
}
