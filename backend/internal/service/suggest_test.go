package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ukapp/internal/dadata"
)

func newSuggestService(t *testing.T, gotCount *int) *Service {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string  `json:"query"`
			Lat   float64 `json:"lat"`
			Lon   float64 `json:"lon"`
			Count int     `json:"count"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		*gotCount = body.Count
		value := body.Query
		if body.Lat != 0 || body.Lon != 0 {
			value = "geo-address"
		}
		_, _ = w.Write([]byte(`{"suggestions":[{"value":"` + value + `","data":{"house_fias_id":"f-1"}}]}`))
	}))
	t.Cleanup(srv.Close)
	dd := dadata.NewClient("k")
	dd.BaseURL = srv.URL
	dd.GeolocateURL = srv.URL
	return New(nil, nil, dd, nil, "testbot")
}

func TestSuggestAddresses_TrimsAndClamps(t *testing.T) {
	var gotCount int
	s := newSuggestService(t, &gotCount)

	got, err := s.SuggestAddresses(context.Background(), "  Казань Баумана 10  ", 99)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Value != "Казань Баумана 10" || got[0].HouseFiasID != "f-1" {
		t.Fatalf("unexpected suggestions %+v", got)
	}
	if gotCount != MaxSuggestCount {
		t.Fatalf("want clamped count %d, got %d", MaxSuggestCount, gotCount)
	}

	if _, err := s.SuggestAddresses(context.Background(), "   ", 5); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestGeolocateAddresses(t *testing.T) {
	var gotCount int
	s := newSuggestService(t, &gotCount)

	got, err := s.GeolocateAddresses(context.Background(), 55.75, 49.1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Value != "geo-address" || got[0].HouseFiasID != "f-1" {
		t.Fatalf("unexpected suggestions %+v", got)
	}
	if gotCount != 3 {
		t.Fatalf("want count 3, got %d", gotCount)
	}

	if _, err := s.GeolocateAddresses(context.Background(), 200, 49.1, 3); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestSuggestAddresses_UpstreamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	dd := dadata.NewClient("k")
	dd.BaseURL = srv.URL
	s := New(nil, nil, dd, nil, "testbot")

	if _, err := s.SuggestAddresses(context.Background(), "казань", 5); err == nil {
		t.Fatal("want upstream error")
	}
}
