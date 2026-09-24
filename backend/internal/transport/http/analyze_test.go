package http

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"ukapp/internal/domain"
)

func analyze(t *testing.T, cls *fakeClassifier, body string) (int, map[string]any) {
	t.Helper()
	srv := newTestServerWith(t, testStore(t), cls)
	bindUser(t, srv, 1, "f-10")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents/analyze", body, 1, "U"))
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func TestAnalyze(t *testing.T) {
	const roof = `{"description":"Крыша течёт после каждого дождя"}`
	code, out := analyze(t, &fakeClassifier{p: domain.Prediction{Category: domain.CategoryBuildingStructure, P: 0.9}}, roof)
	if code != 200 || out["category"] != "BUILDING_STRUCTURE" || out["authority"] != "UK" ||
		out["isUkResponsibility"] != true || out["photoRequired"] != true || out["reasoningText"] == nil {
		t.Fatalf("%d %v", code, out)
	}

	code, out = analyze(t, &fakeClassifier{p: domain.Prediction{Category: domain.CategoryCityTerritory, P: 0.8}}, `{"description":"На дороге за двором огромная яма"}`)
	if code != 200 || out["authority"] != "MUNICIPALITY" || out["isUkResponsibility"] != false {
		t.Fatalf("city: %d %v", code, out)
	}

	if code, _ := analyze(t, &fakeClassifier{p: domain.Prediction{Category: domain.CategoryBuildingStructure, P: 0.5}}, roof); code != 503 {
		t.Fatalf("unsure: %d", code)
	}
	if code, out := analyze(t, &fakeClassifier{err: errors.New("down")}, roof); code != 503 || out["code"] != "model_unavailable" {
		t.Fatalf("down: %d %v", code, out)
	}
	if code, out := analyze(t, &fakeClassifier{}, `{"description":"   коротко   "}`); code != 400 || out["code"] != "validation_failed" {
		t.Fatalf("short: %d %v", code, out)
	}
}
