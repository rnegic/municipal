package layaclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ukapp/internal/domain"
)

func TestClassify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/classify" {
			t.Errorf("path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"category":{"choice":"ELEVATOR","probabilities":{"ELEVATOR":0.8,"WATER_HEAT":0.2}}}`))
	}))
	defer srv.Close()
	p, err := New(srv.URL).Classify(context.Background(), "лифт стоит")
	if err != nil {
		t.Fatal(err)
	}
	if p != (domain.Prediction{Category: domain.CategoryElevator, P: 0.8}) {
		t.Fatalf("%+v", p)
	}
}

func TestClassifyRejectsUnknownLabelAndStatus(t *testing.T) {
	for _, tc := range []struct {
		code int
		body string
	}{
		{200, `{"category":{"choice":"FOO","probabilities":{"FOO":1}}}`},
		{200, `not json`},
		{503, `{"error":"model loading"}`},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.code)
			_, _ = w.Write([]byte(tc.body))
		}))
		if _, err := New(srv.URL).Classify(context.Background(), "x"); err == nil {
			t.Fatalf("code %d body %s: want error", tc.code, tc.body)
		}
		srv.Close()
	}
}
