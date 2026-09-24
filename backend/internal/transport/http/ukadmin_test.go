package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ukapp/internal/repository"
)

const dispatcherPassword = "admin2026"

func seedDispatcher(t *testing.T, s *repository.Store, ukExternalID, inn, validUntil string, ads bool) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(dispatcherPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.DB().ExecContext(context.Background(), `
		WITH org AS (
		  INSERT INTO uk (external_id, name, inn, ogrn, license_number, license_valid_until)
		  VALUES ($1, 'УК '||$1, $2, NULL, '16-'||$2, $3::date)
		  ON CONFLICT (external_id) DO UPDATE SET inn = EXCLUDED.inn, license_number = EXCLUDED.license_number,
		    license_valid_until = EXCLUDED.license_valid_until
		  RETURNING id)
		INSERT INTO app_user (full_name, role, uk_id, position, password_hash, ads_authority)
		SELECT 'Диспетчер '||$1, 'uk_dispatcher', org.id, 'Диспетчер АДС', $4, $5 FROM org`,
		ukExternalID, inn, validUntil, string(hash), ads)
	if err != nil {
		t.Fatal(err)
	}
}

func esiaLogin(t *testing.T, srv http.Handler, inn, password string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest("POST", "/api/auth/esia-mock", stringsReader(`{"inn":"`+inn+`","password":"`+password+`"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	return w
}

func dispatcherToken(t *testing.T, srv http.Handler, inn string) string {
	t.Helper()
	w := esiaLogin(t, srv, inn, dispatcherPassword)
	if w.Code != 200 {
		t.Fatalf("login: %d %s", w.Code, w.Body)
	}
	var out struct{ Token string }
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return out.Token
}

func bearerReq(method, path, body, token string) *http.Request {
	r := httptest.NewRequest(method, path, stringsReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	r.Header.Set("Authorization", "Bearer "+token)
	return r
}

func serve(srv http.Handler, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	return w
}

func TestEsiaMockLogin(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	seedDispatcher(t, s, "uk-1", "1655000003", "2099-12-31", true)
	seedDispatcher(t, s, "uk-expired", "7707083893", "2020-01-01", true)

	if w := esiaLogin(t, srv, "1655000000", dispatcherPassword); w.Code != 400 {
		t.Fatalf("bad checksum: want 400 got %d", w.Code)
	}
	unknown := esiaLogin(t, srv, "7736050003", dispatcherPassword)
	wrongPass := esiaLogin(t, srv, "1655000003", "nope")
	if unknown.Code != 403 || wrongPass.Code != 403 || unknown.Body.String() != wrongPass.Body.String() {
		t.Fatalf("unknown org and wrong password must be the same 403: %d %s / %d %s",
			unknown.Code, unknown.Body, wrongPass.Code, wrongPass.Body)
	}
	if w := esiaLogin(t, srv, "7707083893", dispatcherPassword); w.Code != 403 {
		t.Fatalf("expired license: want 403 got %d", w.Code)
	}

	w := esiaLogin(t, srv, "1655000003", dispatcherPassword)
	if w.Code != 200 {
		t.Fatalf("login: %d %s", w.Code, w.Body)
	}
	var raw map[string]map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &raw)
	org := raw["organization"]
	if _, ok := org["ogrn"]; !ok || org["ogrn"] != nil || org["inn"] != "1655000003" || raw["user"]["role"] != "uk_dispatcher" {
		t.Fatalf("session shape: %s", w.Body)
	}
	var sess struct {
		Token, AuthMethod string
		ExpiresAt         time.Time
	}
	_ = json.Unmarshal(w.Body.Bytes(), &sess)
	if sess.AuthMethod != "esia_mock" || sess.Token == "" || time.Until(sess.ExpiresAt) < time.Hour {
		t.Fatalf("session: %+v", sess)
	}

	var me struct {
		User  struct{ Role string }
		House *struct{}
	}
	mw := serve(srv, bearerReq("GET", "/api/me", "", sess.Token))
	_ = json.Unmarshal(mw.Body.Bytes(), &me)
	if mw.Code != 200 || me.User.Role != "uk_dispatcher" || me.House != nil {
		t.Fatalf("me under bearer: %d %s", mw.Code, mw.Body)
	}
}

func TestEsiaMockLogin_RateLimited(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	seedDispatcher(t, s, "uk-1", "1655000003", "2099-12-31", true)
	for i := 0; i < 5; i++ {
		if w := esiaLogin(t, srv, "1655000003", "wrong"); w.Code != 403 {
			t.Fatalf("attempt %d: want 403 got %d", i, w.Code)
		}
	}
	if w := esiaLogin(t, srv, "1655000003", dispatcherPassword); w.Code != 429 {
		t.Fatalf("6th attempt: want 429 got %d", w.Code)
	}
}

func TestEsiaMockLogin_AuditsMaxUser(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	seedDispatcher(t, s, "uk-1", "1655000003", "2099-12-31", true)
	r := authedReq(t, "POST", "/api/auth/esia-mock", `{"inn":"1655000003","password":"`+dispatcherPassword+`"}`, 555, "U")
	if w := serve(srv, r); w.Code != 200 {
		t.Fatalf("login with tma: %d %s", w.Code, w.Body)
	}
	var maxID int64
	if err := s.DB().QueryRowContext(context.Background(),
		`SELECT max_user_id FROM uk_login_attempt WHERE success`).Scan(&maxID); err != nil || maxID != 555 {
		t.Fatalf("audit: %d %v", maxID, err)
	}
}

func TestEsiaMockLogin_RequiresAdsAuthority(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	seedDispatcher(t, s, "uk-1", "1655000003", "2099-12-31", false)
	if w := esiaLogin(t, srv, "1655000003", dispatcherPassword); w.Code != 403 {
		t.Fatalf("no ADS authority: want 403 got %d", w.Code)
	}
}

func TestRoleGuards(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	bindUser(t, srv, 1, "f-10")
	inc, _ := createIncident(t, srv, 1, `{"description":"нет воды с самого утра","category":"WATER_HEAT","photoUrls":[]}`)
	seedDispatcher(t, s, "uk-1", "1655000003", "2099-12-31", true)
	token := dispatcherToken(t, srv, "1655000003")

	for _, c := range []struct {
		name string
		r    *http.Request
		want int
	}{
		{"resident patches status", authedReq(t, "PATCH", "/api/incidents/"+inc.Id+"/status", `{"status":"in_progress"}`, 1, "U"), 403},
		{"resident reads queue", authedReq(t, "GET", "/api/uk/queue", "", 1, "U"), 403},
		{"resident creates event", authedReq(t, "POST", "/api/uk/events", `{}`, 1, "U"), 403},
		{"dispatcher joins", bearerReq("POST", "/api/incidents/"+inc.Id+"/join", "", token), 403},
		{"dispatcher binds house", bearerReq("POST", "/api/houses/bind", `{"address":"f-10"}`, token), 403},
		{"bad bearer", bearerReq("GET", "/api/me", "", token+"x"), 401},
		{"no auth", httptest.NewRequest("GET", "/api/uk/queue", nil), 401},
		{"dispatcher reads queue", bearerReq("GET", "/api/uk/queue", "", token), 200},
	} {
		if w := serve(srv, c.r); w.Code != c.want {
			t.Errorf("%s: want %d got %d %s", c.name, c.want, w.Code, w.Body)
		}
	}
}

func TestUkQueueAndIsolation(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	h10 := bindUser(t, srv, 1, "f-10")
	bindUser(t, srv, 3, "f-12")
	seedDispatcher(t, s, "uk-1", "1655000003", "2099-12-31", true)
	seedDispatcher(t, s, "uk-2", "7707083893", "2099-12-31", true)
	if _, err := s.DB().ExecContext(context.Background(),
		`UPDATE house SET uk_id = (SELECT id FROM uk WHERE external_id = 'uk-2') WHERE house_fias_id = 'f-12'`); err != nil {
		t.Fatal(err)
	}
	own, _ := createIncident(t, srv, 1, `{"description":"нет воды с самого утра","category":"WATER_HEAT","photoUrls":[]}`)
	foreign, _ := createIncident(t, srv, 3, `{"description":"в подъезде не горит свет","category":"ELECTRICITY","photoUrls":[]}`)
	incID, _ := parseID("inc_", own.Id)
	setVerifying(t, s, incID)
	confirm := authedReq(t, "POST", "/api/incidents/"+own.Id+"/confirm", "", 1, "U")
	if w := serve(srv, confirm); w.Code != 200 {
		t.Fatalf("confirm: %d %s", w.Code, w.Body)
	}
	token := dispatcherToken(t, srv, "1655000003")

	var q struct {
		Items []struct {
			Id, HouseId, HouseAddress, ReporterName, Status string
			AffectedCount, ConfirmedCount                   int
			DueAt                                           *string
		}
		Total, Offset, Limit int
	}
	w := serve(srv, bearerReq("GET", "/api/uk/queue?limit=10", "", token))
	_ = json.Unmarshal(w.Body.Bytes(), &q)
	if w.Code != 200 || q.Total != 1 || len(q.Items) != 1 || q.Limit != 10 {
		t.Fatalf("queue: %d %s", w.Code, w.Body)
	}
	it := q.Items[0]
	if it.Id != own.Id || it.HouseId != h10 || it.HouseAddress != "f-10" || it.ReporterName != "U T" ||
		it.AffectedCount != 1 || it.ConfirmedCount != 1 || it.DueAt == nil {
		t.Fatalf("queue item: %+v", it)
	}

	if w := serve(srv, bearerReq("GET", "/api/incidents/"+foreign.Id, "", token)); w.Code != 404 {
		t.Fatalf("foreign incident card: want 404 got %d", w.Code)
	}
	if w := serve(srv, bearerReq("PATCH", "/api/incidents/"+foreign.Id+"/status", `{"status":"in_progress"}`, token)); w.Code != 404 {
		t.Fatalf("foreign status change: want 404 got %d", w.Code)
	}
	if w := serve(srv, bearerReq("GET", "/api/incidents/"+own.Id, "", token)); w.Code != 200 {
		t.Fatalf("own incident card: want 200 got %d", w.Code)
	}
}

func TestEvents(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	h10 := bindUser(t, srv, 1, "f-10")
	h12 := bindUser(t, srv, 3, "f-12")
	seedDispatcher(t, s, "uk-1", "1655000003", "2099-12-31", true)
	seedDispatcher(t, s, "uk-2", "7707083893", "2099-12-31", true)
	if _, err := s.DB().ExecContext(context.Background(),
		`UPDATE house SET uk_id = (SELECT id FROM uk WHERE external_id = 'uk-2') WHERE house_fias_id = 'f-12'`); err != nil {
		t.Fatal(err)
	}
	token := dispatcherToken(t, srv, "1655000003")
	from := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	to := time.Now().Add(3 * time.Hour).UTC().Format(time.RFC3339)
	body := func(house string) string {
		return `{"houseId":"` + house + `","reason":"Отключение ГВС","responsible":"Бригада 2","riser":"3","scheduledFrom":"` + from + `","scheduledTo":"` + to + `"}`
	}

	w := serve(srv, bearerReq("POST", "/api/uk/events", body(h10), token))
	if w.Code != 201 {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	if w := serve(srv, bearerReq("POST", "/api/uk/events", body(h12), token)); w.Code != 404 {
		t.Fatalf("foreign house: want 404 got %d", w.Code)
	}
	bad := `{"houseId":"` + h10 + `","reason":"x","responsible":"y","scheduledFrom":"` + to + `","scheduledTo":"` + from + `"}`
	if w := serve(srv, bearerReq("POST", "/api/uk/events", bad, token)); w.Code != 400 {
		t.Fatalf("reversed interval: want 400 got %d", w.Code)
	}

	var list struct {
		Items []map[string]any
	}
	if c := getJSON(t, srv, "/api/events?houseId="+h10, 1, &list); c != 200 || len(list.Items) != 1 {
		t.Fatalf("resident events: %d %+v", c, list)
	}
	ev := list.Items[0]
	if ev["status"] != "planned" || ev["riser"] != "3" || ev["entrance"] != nil || ev["resolvedAt"] != nil {
		t.Fatalf("event: %+v", ev)
	}
	if _, ok := ev["entrance"]; !ok {
		t.Fatal("nullable entrance key must be present")
	}
	if c := getJSON(t, srv, "/api/events?houseId="+h10, 3, nil); c != 404 {
		t.Fatalf("other resident: want 404 got %d", c)
	}
	if w := serve(srv, bearerReq("GET", "/api/events?houseId="+h10, "", token)); w.Code != 200 {
		t.Fatalf("dispatcher events: %d", w.Code)
	}
}

func TestUnbindHouse(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	bindUser(t, srv, 1, "f-10")
	for i := 0; i < 2; i++ {
		if w := serve(srv, authedReq(t, "DELETE", "/api/houses/bind", "", 1, "U")); w.Code != 204 {
			t.Fatalf("unbind: %d %s", w.Code, w.Body)
		}
	}
	var me struct{ House *struct{} }
	getJSON(t, srv, "/api/me", 1, &me)
	if me.House != nil {
		t.Fatal("house must be null after unbind")
	}
	pre := httptest.NewRequest("OPTIONS", "/api/houses/bind", nil)
	if w := serve(srv, pre); w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PATCH, DELETE, OPTIONS" {
		t.Fatalf("cors: %q", w.Header().Get("Access-Control-Allow-Methods"))
	}
}
