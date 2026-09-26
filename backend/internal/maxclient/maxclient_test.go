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

func TestClient_UpdatesParsesBotStartedAndMarker(t *testing.T) {
	var gotMarker string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMarker = r.URL.Query().Get("marker")
		_, _ = w.Write([]byte(`{"updates":[{"update_type":"bot_started","timestamp":1,"chat_id":5,"user":{"user_id":42,"first_name":"Анна","name":"Анна"},"payload":null}],"marker":7}`))
	}))
	defer srv.Close()

	c := NewClient("tok")
	c.BaseURL = srv.URL
	prev := int64(3)
	ups, next, err := c.Updates(context.Background(), &prev, "bot_started")
	if err != nil {
		t.Fatal(err)
	}
	if gotMarker != "3" || next == nil || *next != 7 || len(ups) != 1 || ups[0].UpdateType != "bot_started" || ups[0].User.UserID != 42 || ups[0].User.FirstName != "Анна" {
		t.Fatalf("marker=%s next=%v ups=%+v", gotMarker, next, ups)
	}
}
