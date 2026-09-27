package http

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

type meResponse struct {
	User struct {
		Id       string
		FullName string
		Role     string
	}
	House *struct {
		Id      string
		Address string
		Uk      struct {
			Name           *string
			Phone          *string
			EmergencyPhone *string
		}
	}
}

func TestMe_CreatesOnFirstCall(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/me", "", 500, "Anna"))
	if w.Code != 200 {
		t.Fatalf("code %d body %s", w.Code, w.Body)
	}
	var me meResponse
	if err := json.Unmarshal(w.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me.User.FullName != "Anna T" || me.User.Role != "resident" || me.House != nil {
		t.Fatalf("bad me %+v", me)
	}

	srv.ServeHTTP(httptest.NewRecorder(), authedReq(t, "GET", "/api/me", "", 500, "Anna"))
	var n int
	if err := s.DB().QueryRowContext(t.Context(), `SELECT count(*) FROM app_user`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 user, got %d", n)
	}
}

func TestMe_Unauthorized(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest("GET", "/api/me", nil))
	if w.Code != 401 {
		t.Fatalf("code %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"code":"unauthorized"`) {
		t.Fatalf("expected JSON error, got %s", w.Body)
	}
}

func TestBindHouse(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)

	h1 := bindUser(t, srv, 501, "abc-123")
	if h1 == "" {
		t.Fatal("house id not set")
	}
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/me", "", 501, "Ivan"))
	var me meResponse
	_ = json.Unmarshal(w.Body.Bytes(), &me)
	if me.House == nil || me.House.Id != h1 {
		t.Fatalf("house not bound in profile: %+v", me)
	}
	if me.House.Uk.Name == nil || *me.House.Uk.Name != "Демо УК" ||
		me.House.Uk.Phone == nil || *me.House.Uk.Phone != "+7 (843) 200-00-00" ||
		me.House.Uk.EmergencyPhone == nil || *me.House.Uk.EmergencyPhone != "+7 (843) 200-01-01" {
		t.Fatalf("house uk contacts missing in profile: %+v", me.House.Uk)
	}

	h2 := bindUser(t, srv, 502, "abc-123")
	if h2 != h1 {
		t.Fatalf("same address must map to same house: %v vs %v", h1, h2)
	}

	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/houses/bind", `{"address":""}`, 501, "Ivan"))
	if w.Code != 400 {
		t.Fatalf("code %d body %s", w.Code, w.Body)
	}
}

func TestBindHouse_NotServedByUk(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/houses/bind", `{"address":"unknown"}`, 777, "U"))
	if w.Code != 422 || !strings.Contains(w.Body.String(), `"code":"business_rule_failed"`) {
		t.Fatalf("want 422 business_rule_failed, got %d %s", w.Code, w.Body)
	}
}

func TestBindHouse_CreatesUkFromProvider(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	bindUser(t, srv, 778, "f-42")
	var ext, name, houseExt string
	err := s.DB().QueryRowContext(t.Context(),
		`SELECT uk.external_id, uk.name, house.external_id FROM house JOIN uk ON uk.id = house.uk_id WHERE house.house_fias_id='f-42'`).
		Scan(&ext, &name, &houseExt)
	if err != nil {
		t.Fatal(err)
	}
	if ext != "uk-1" || name != "Демо УК" || houseExt != "h-f-42" {
		t.Fatalf("got uk %s/%s house %s", ext, name, houseExt)
	}
}

func TestSuggestAddresses(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/houses/suggest?query=f-10&count=3", "", 900, "Resident"))
	if w.Code != 200 {
		t.Fatalf("code %d body %s", w.Code, w.Body)
	}
	var resp struct {
		Suggestions []struct {
			Value       string
			HouseFiasId string
		}
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Suggestions) != 1 || resp.Suggestions[0].Value != "f-10" || resp.Suggestions[0].HouseFiasId != "f-10" {
		t.Fatalf("unexpected suggestions %+v", resp)
	}

	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/houses/suggest?query=", "", 900, "Resident"))
	if w.Code != 400 || !strings.Contains(w.Body.String(), `"code":"validation_failed"`) {
		t.Fatalf("want 400 validation_failed, got %d %s", w.Code, w.Body)
	}

	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/houses/suggest?lat=55.75&lon=49.1", "", 900, "Resident"))
	if w.Code != 200 {
		t.Fatalf("geolocate: code %d body %s", w.Code, w.Body)
	}
	resp = struct {
		Suggestions []struct {
			Value       string
			HouseFiasId string
		}
	}{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Suggestions) != 1 || resp.Suggestions[0].Value != "geo-address" {
		t.Fatalf("unexpected geolocate suggestions %+v", resp)
	}

	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/houses/suggest?lat=200&lon=49.1", "", 900, "Resident"))
	if w.Code != 400 || !strings.Contains(w.Body.String(), `"code":"validation_failed"`) {
		t.Fatalf("want 400 validation_failed, got %d %s", w.Code, w.Body)
	}

	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/houses/suggest", "", 900, "Resident"))
	if w.Code != 400 || !strings.Contains(w.Body.String(), `"code":"validation_failed"`) {
		t.Fatalf("want 400 validation_failed, got %d %s", w.Code, w.Body)
	}

	w = httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest("GET", "/api/houses/suggest?query=f-10", nil))
	if w.Code != 401 {
		t.Fatalf("want 401, got %d", w.Code)
	}
}
