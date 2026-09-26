package dedupclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ukapp/internal/domain"
)

func riser(s string) *string { return &s }

var (
	req   = domain.OpenIncident{Title: "Нет горячей воды", Description: "с утра", Category: domain.CategoryBuildingStructure, Riser: riser("5 этаж"), Severity: domain.SeverityCritical}
	cands = []domain.OpenIncident{{ID: 7, Title: "Горячей нет", Description: "стояк 3", Category: domain.CategoryWaterHeat, Riser: riser("3"), Severity: domain.SeverityCritical, CreatedAt: time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC)}}
)

func TestMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/match" {
			t.Errorf("path %s", r.URL.Path)
		}
		var body struct {
			Request    map[string]any   `json:"request"`
			Candidates []map[string]any `json:"candidates"`
			TopK       int              `json:"top_k"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		c := body.Candidates[0]
		if body.TopK != 1 || c["id"] != float64(7) || c["riser"] != "3" || c["entrance"] != nil || c["severity"] != "critical" ||
			body.Request["title"] != "Нет горячей воды" || body.Request["riser"] != nil || c["created_at"] != "2026-09-25T10:00:00Z" {
			t.Errorf("body %+v", body)
		}
		_, _ = w.Write([]byte(`{"match":7,"duplicate_probability":0.97,"model_version":"m@1","threshold":0.9,"scores":[]}`))
	}))
	defer srv.Close()
	m, err := New(srv.URL, time.Second).Match(context.Background(), req, cands)
	if err != nil || m != (domain.DedupMatch{IncidentID: 7, P: 0.97, ModelVersion: "m@1"}) {
		t.Fatalf("%+v %v", m, err)
	}
}

func TestMatchNone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"match":null,"duplicate_probability":0.2,"model_version":"m@1","threshold":0.9,"scores":[]}`))
	}))
	defer srv.Close()
	m, err := New(srv.URL, time.Second).Match(context.Background(), req, cands)
	if err != nil || m.IncidentID != 0 || m.ModelVersion != "m@1" {
		t.Fatalf("%+v %v", m, err)
	}
}

func TestMatchErrors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		code  int
		body  string
		delay time.Duration
	}{
		{"status", 503, `{"error":"loading"}`, 0},
		{"json", 200, `not json`, 0},
		{"no version", 200, `{"match":null}`, 0},
		{"foreign id", 200, `{"match":99,"duplicate_probability":0.99,"model_version":"m@1"}`, 0},
		{"timeout", 200, `{"match":null,"model_version":"m@1"}`, 300 * time.Millisecond},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(tc.delay)
			w.WriteHeader(tc.code)
			_, _ = w.Write([]byte(tc.body))
		}))
		start := time.Now()
		_, err := New(srv.URL, 100*time.Millisecond).Match(context.Background(), req, cands)
		if err == nil || time.Since(start) > 250*time.Millisecond {
			t.Errorf("%s: err=%v after %s", tc.name, err, time.Since(start))
		}
		srv.Close()
	}
}
