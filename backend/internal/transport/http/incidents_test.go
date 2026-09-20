package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type incidentResp struct {
	Id            string
	HouseId       string
	Title         string
	Description   string
	Severity      string
	Status        string
	AffectedCount int
	JoinedByMe    bool
	ConfirmedByMe bool
}

func createIncident(t *testing.T, srv http.Handler, maxID int64, body string) (incidentResp, int) {
	t.Helper()
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents", body, maxID, "U"))
	var out incidentResp
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return out, w.Code
}

func TestIncidents_DedupAndJoin(t *testing.T) {
	s := testStore(t)
	seedUK(t, s)
	srv := newTestServer(t, s)
	h10 := bindUser(t, srv, 1, "f-10")
	bindUser(t, srv, 2, "f-10")
	bindUser(t, srv, 3, "f-12")

	inc1, code := createIncident(t, srv, 1, `{"title":"Нет воды","description":"нет воды","severity":"critical"}`)
	if code != 201 {
		t.Fatalf("create: %d", code)
	}
	inc2, code := createIncident(t, srv, 2, `{"title":"Нет воды","description":"течёт кипяток","severity":"critical"}`)
	if code != 200 || inc2.Id != inc1.Id {
		t.Fatalf("same house+title must merge: code=%d id2=%s", code, inc2.Id)
	}
	// dedup'd report shows up in the joiner's "мои заявки"
	w0 := httptest.NewRecorder()
	srv.ServeHTTP(w0, authedReq(t, "GET", "/api/houses/"+h10+"/requests", "", 2, "U"))
	var mine struct{ Items []struct{ Id string } }
	_ = json.Unmarshal(w0.Body.Bytes(), &mine)
	if w0.Code != 200 || len(mine.Items) != 1 || mine.Items[0].Id != inc1.Id {
		t.Fatalf("joiner's requests: %d %s", w0.Code, w0.Body)
	}
	inc3, code := createIncident(t, srv, 2, `{"title":"Лифт не едет","description":"лифт","severity":"warning"}`)
	if code != 201 || inc3.Id == inc1.Id {
		t.Fatal("different title must not merge")
	}
	inc4, code := createIncident(t, srv, 3, `{"title":"Нет воды","description":"нет воды","severity":"critical"}`)
	if code != 201 || inc4.Id == inc1.Id {
		t.Fatal("different house must not merge")
	}

	// active list for house f-10: critical first, then createdAt desc
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/houses/"+h10+"/incidents?status=active", "", 1, "U"))
	if w.Code != 200 {
		t.Fatalf("list: %d %s", w.Code, w.Body)
	}
	var list struct{ Items []incidentResp }
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list.Items) != 2 {
		t.Fatalf("want 2 incidents, got %d: %s", len(list.Items), w.Body)
	}
	if list.Items[0].Severity != "critical" {
		t.Fatalf("critical must come first: %+v", list.Items)
	}
	for _, inc := range list.Items {
		switch inc.Id {
		case inc1.Id:
			if inc.AffectedCount != 2 || !inc.JoinedByMe {
				t.Fatalf("water incident: %+v", inc)
			}
		case inc3.Id:
			if inc.AffectedCount != 1 || inc.JoinedByMe {
				t.Fatalf("elevator incident: %+v", inc)
			}
		}
	}

	// join ("у меня тоже") is idempotent
	for i := 0; i < 2; i++ {
		w = httptest.NewRecorder()
		srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents/"+inc3.Id+"/join", "", 1, "U"))
		if w.Code != 200 {
			t.Fatalf("join: %d %s", w.Code, w.Body)
		}
		var jr struct {
			AffectedCount int
			Joined        bool
		}
		_ = json.Unmarshal(w.Body.Bytes(), &jr)
		if jr.AffectedCount != 2 || !jr.Joined {
			t.Fatalf("bad join response: %+v", jr)
		}
	}

	// join of a missing incident → 404
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents/inc_999999/join", "", 1, "U"))
	if w.Code != 404 {
		t.Fatalf("join missing: %d %s", w.Code, w.Body)
	}
}

func TestIncidents_RequiresHouseAndValidInput(t *testing.T) {
	s := testStore(t)
	seedUK(t, s)
	srv := newTestServer(t, s)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents", `{"title":"x","description":"x","severity":"critical"}`, 9, "U"))
	if w.Code != 404 {
		t.Fatalf("no house: want 404 got %d", w.Code)
	}
	bindUser(t, srv, 9, "f-1")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents", `{"title":"x","description":"x","severity":"urgent"}`, 9, "U"))
	if w.Code != 400 {
		t.Fatalf("bad severity: want 400 got %d", w.Code)
	}
}

func TestGetIncident(t *testing.T) {
	s := testStore(t)
	seedUK(t, s)
	srv := newTestServer(t, s)
	bindUser(t, srv, 1, "f-10")
	inc, _ := createIncident(t, srv, 1, `{"title":"Нет воды","description":"x","severity":"critical","entrance":"2","riser":"7"}`)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/incidents/"+inc.Id, "", 1, "U"))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"entrance":"2"`) || !strings.Contains(w.Body.String(), `"riser":"7"`) {
		t.Fatalf("get: %d %s", w.Code, w.Body)
	}
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/incidents/inc_999999", "", 1, "U"))
	if w.Code != 404 {
		t.Fatalf("missing: %d", w.Code)
	}
}

func TestConfirmIncident_RequiresVerifyingAndCloses(t *testing.T) {
	s := testStore(t)
	seedUK(t, s)
	srv := newTestServer(t, s)
	bindUser(t, srv, 1, "f-10")
	bindUser(t, srv, 2, "f-10")
	inc, _ := createIncident(t, srv, 1, `{"title":"Нет воды","description":"x","severity":"critical"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents/"+inc.Id+"/join", "", 2, "U"))
	if w.Code != 200 {
		t.Fatalf("join: %d", w.Code)
	}

	// not verifying yet → 422
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents/"+inc.Id+"/confirm", "", 1, "U"))
	if w.Code != 422 {
		t.Fatalf("premature confirm: want 422 got %d %s", w.Code, w.Body)
	}

	incID, _ := parseID("inc_", inc.Id)
	makeDispatcher(t, srv, s, 100)
	if c := setStatus(t, srv, 100, inc.Id, "in_progress"); c != 200 {
		t.Fatalf("→ in_progress: %d", c)
	}
	if c := setStatus(t, srv, 100, inc.Id, "verifying"); c != 200 {
		t.Fatalf("→ verifying: %d", c)
	}

	// 1 of 2 confirmations (50%, but under MinConfirmations=2) → stays verifying
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents/"+inc.Id+"/confirm", "", 1, "U"))
	if w.Code != 200 {
		t.Fatalf("confirm 1: %d %s", w.Code, w.Body)
	}
	var cr struct{ Status string }
	_ = json.Unmarshal(w.Body.Bytes(), &cr)
	if cr.Status != "verifying" {
		t.Fatalf("want still verifying, got %s", cr.Status)
	}

	// 2 of 2 confirmations → done
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents/"+inc.Id+"/confirm", "", 2, "U"))
	if w.Code != 200 {
		t.Fatalf("confirm 2: %d %s", w.Code, w.Body)
	}
	_ = json.Unmarshal(w.Body.Bytes(), &cr)
	if cr.Status != "done" {
		t.Fatalf("want done, got %s", cr.Status)
	}
	if st := incidentStatus(t, s, incID); st != "done" {
		t.Fatalf("db status: %s", st)
	}

	// idempotent repeat
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents/"+inc.Id+"/confirm", "", 1, "U"))
	if w.Code != 422 {
		t.Fatalf("confirm on a done incident: want 422 got %d", w.Code)
	}
}

func TestListHouseRequests_Paginated(t *testing.T) {
	s := testStore(t)
	seedUK(t, s)
	srv := newTestServer(t, s)
	h10 := bindUser(t, srv, 1, "f-10")
	createIncident(t, srv, 1, `{"title":"A","description":"x","severity":"warning"}`)
	createIncident(t, srv, 1, `{"title":"B","description":"x","severity":"warning"}`)
	createIncident(t, srv, 1, `{"title":"C","description":"x","severity":"warning"}`)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", "/api/houses/"+h10+"/requests?offset=0&limit=2", "", 1, "U"))
	if w.Code != 200 {
		t.Fatalf("requests: %d %s", w.Code, w.Body)
	}
	var page struct {
		Items []struct {
			Id            string
			Title         string
			Status        string
			CreatedAt     time.Time
			DueAt         *time.Time
			ConfirmedByMe *bool
		}
		Total  int
		Offset int
		Limit  int
	}
	_ = json.Unmarshal(w.Body.Bytes(), &page)
	if page.Total != 3 || len(page.Items) != 2 || page.Items[0].Title != "C" {
		t.Fatalf("bad page: %+v", page)
	}
	for _, it := range page.Items {
		// id is the incident id (usable with /confirm), dueAt = createdAt + SLA(warning)=24h
		if !strings.HasPrefix(it.Id, "inc_") || it.DueAt == nil || !it.DueAt.Equal(it.CreatedAt.Add(24*time.Hour)) || it.ConfirmedByMe == nil {
			t.Fatalf("bad item: %+v", it)
		}
	}
}
