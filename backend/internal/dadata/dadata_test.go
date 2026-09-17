package dadata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Suggest(t *testing.T) {
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token k" {
			t.Errorf("bad auth header %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"suggestions":[{"value":"г Казань, ул Баумана, д 10","data":{"house_fias_id":"abc-123"}},{"value":"без дома","data":{"house_fias_id":null}}]}`))
	}))
	defer fake.Close()
	c := NewClient("k")
	c.BaseURL = fake.URL
	got, err := c.Suggest(context.Background(), "казань баумана 10")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].HouseFiasID != "abc-123" {
		t.Fatalf("want 1 suggestion with house, got %+v", got)
	}
}
