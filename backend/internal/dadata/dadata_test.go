package dadata

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeSuggestServer(t *testing.T, gotCount *int) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token k" {
			t.Errorf("bad auth header %q", r.Header.Get("Authorization"))
		}
		var body struct {
			Count int `json:"count"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		*gotCount = body.Count
		_, _ = w.Write([]byte(`{"suggestions":[{"value":"г Казань, ул Баумана, д 10","data":{"house_fias_id":"abc-123"}},{"value":"без дома","data":{"house_fias_id":null}}]}`))
	}))
	t.Cleanup(srv.Close)
	c := NewClient("k")
	c.BaseURL = srv.URL
	return c
}

func TestClient_Suggest(t *testing.T) {
	var gotCount int
	c := fakeSuggestServer(t, &gotCount)

	got, err := c.Suggest(context.Background(), "казань баумана 10", 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].HouseFiasID != "abc-123" {
		t.Fatalf("want 1 suggestion with house, got %+v", got)
	}
	if gotCount != 7 {
		t.Fatalf("want count 7 in request, got %d", gotCount)
	}
}

func TestClient_Suggest_DefaultCount(t *testing.T) {
	var gotCount int
	c := fakeSuggestServer(t, &gotCount)

	if _, err := c.Suggest(context.Background(), "казань", 0); err != nil {
		t.Fatal(err)
	}
	if gotCount != defaultSuggestCount {
		t.Fatalf("want default count %d, got %d", defaultSuggestCount, gotCount)
	}
}
