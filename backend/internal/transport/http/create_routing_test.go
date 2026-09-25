package http

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"ukapp/internal/domain"
	"ukapp/internal/repository"
)

func routingSource(t *testing.T, s *repository.Store, wireID string) (source string, predicted *string) {
	t.Helper()
	id, _ := parseID("inc_", wireID)
	if err := s.DB().QueryRowContext(context.Background(),
		`SELECT routing_source, category_predicted FROM incident WHERE id=$1`, id).Scan(&source, &predicted); err != nil {
		t.Fatal(err)
	}
	return source, predicted
}

func TestCreate_RoutingRules(t *testing.T) {
	s := testStore(t)
	cls := &fakeClassifier{err: errors.New("down")}
	srv := newTestServerWith(t, s, cls)
	bindUser(t, srv, 1, "f-10")
	_, url := uploadStaged(t, srv, 1, pngBytes)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents", `{"description":"открыт люк на проезжей части","category":"CITY_TERRITORY","photoUrls":["`+url+`"]}`, 1, "U"))
	if w.Code != 422 || !strings.Contains(w.Body.String(), "муниципалитет") {
		t.Fatalf("non-UK category must be 422: %d %s", w.Code, w.Body)
	}

	if _, code := createIncident(t, srv, 1, `{"description":"мусор во дворе неделю","category":"CLEANING_YARD","photoUrls":[]}`); code != 422 {
		t.Fatalf("photo required: %d", code)
	}
	if _, code := createIncident(t, srv, 1, `{"description":"мусор во дворе неделю","category":"CLEANING_YARD","photoUrls":["https://evil/x.jpg"]}`); code != 400 {
		t.Fatalf("foreign url: %d", code)
	}
	if _, code := createIncident(t, srv, 1, `{"description":"мусор во дворе неделю","category":"CLEANING_YARD","floorZone":"`+strings.Repeat("я", 121)+`","photoUrls":[]}`); code != 400 {
		t.Fatalf("floorZone > 120: %d", code)
	}
	inc, code := createIncident(t, srv, 1, `{"description":"мусор во дворе неделю","category":"CLEANING_YARD","photoUrls":["`+url+`","`+url+`"]}`)
	if code != 201 {
		t.Fatalf("model down → manual create: %d", code)
	}
	if src, pred := routingSource(t, s, inc.Id); src != "manual" || pred != nil {
		t.Fatalf("routing_source=%q predicted=%v", src, pred)
	}
	if _, code := createIncident(t, srv, 1, `{"description":"лифт стоит третий день","category":"ELEVATOR","photoUrls":["`+url+`"]}`); code != 422 {
		t.Fatalf("already bound photo: %d", code)
	}
}

func TestCreate_ForeignPhotoAndDedupBinding(t *testing.T) {
	s := testStore(t)
	cls := &fakeClassifier{p: domain.Prediction{Category: domain.CategoryElevator, P: 0.9}}
	srv := newTestServerWith(t, s, cls)
	bindUser(t, srv, 1, "f-10")
	bindUser(t, srv, 2, "f-10")
	_, mine := uploadStaged(t, srv, 1, pngBytes)
	_, theirs := uploadStaged(t, srv, 2, pngBytes)

	if _, code := createIncident(t, srv, 1, `{"description":"лифт стоит третий день","category":"ELEVATOR","floorZone":"1 подъезд","photoUrls":["`+theirs+`"]}`); code != 422 {
		t.Fatalf("foreign photo: %d", code)
	}
	first, code := createIncident(t, srv, 1, `{"title":"Лифт не работает","description":"лифт стоит третий день","category":"ELEVATOR","floorZone":"1 подъезд","photoUrls":["`+mine+`"]}`)
	if code != 201 {
		t.Fatalf("create: %d", code)
	}
	if src, pred := routingSource(t, s, first.Id); src != "auto" || pred == nil || *pred != "ELEVATOR" {
		t.Fatalf("routing_source=%q predicted=%v", src, pred)
	}
	second, code := createIncident(t, srv, 2, `{"description":"лифт не едет совсем","category":"ELEVATOR","floorZone":"1 подъезд","photoUrls":["`+theirs+`"]}`)
	if code != 200 || second.Id != first.Id {
		t.Fatalf("dedup: %d %s %s", code, second.Id, first.Id)
	}
	var full struct {
		Title    string
		Severity string
		Category string
		Riser    string
		Photos   []struct{ Url string }
	}
	getJSON(t, srv, "/api/incidents/"+first.Id, 1, &full)
	if full.Category != "ELEVATOR" || full.Title != "Лифт не работает" || full.Severity != "critical" || full.Riser != "1 подъезд" || len(full.Photos) != 2 {
		t.Fatalf("%+v", full)
	}

	cls.p = domain.Prediction{Category: domain.CategoryElevator, P: 0.5}
	third, code := createIncident(t, srv, 1, `{"description":"нет воды с самого утра","category":"WATER_HEAT","photoUrls":[]}`)
	if code != 201 {
		t.Fatalf("unsure create: %d", code)
	}
	if src, pred := routingSource(t, s, third.Id); src != "manual" || pred == nil || *pred != "ELEVATOR" {
		t.Fatalf("unsure: routing_source=%q predicted=%v", src, pred)
	}
}
