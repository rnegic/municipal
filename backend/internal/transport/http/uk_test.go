package http

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func getJSON(t *testing.T, srv http.Handler, path string, maxID int64, out any) int {
	t.Helper()
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "GET", path, "", maxID, "U"))
	if out != nil {
		_ = json.Unmarshal(w.Body.Bytes(), out)
	}
	return w.Code
}

func TestSetIncidentStatus_Pipeline(t *testing.T) {
	s := testStore(t)
	seedUK(t, s)
	srv := newTestServer(t, s)
	bindUser(t, srv, 1, "f-10")
	bindUser(t, srv, 2, "f-10")
	makeDispatcher(t, srv, s, 100)
	inc, _ := createIncident(t, srv, 1, `{"title":"Нет воды","description":"x","severity":"critical"}`)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents/"+inc.Id+"/join", "", 2, "U"))

	if c := setStatus(t, srv, 1, inc.Id, "in_progress"); c != 403 {
		t.Fatalf("resident: want 403 got %d", c)
	}
	if c := setStatus(t, srv, 100, inc.Id, "verifying"); c != 422 {
		t.Fatalf("skip a step: want 422 got %d", c)
	}
	if c := setStatus(t, srv, 100, inc.Id, "urgent"); c != 400 {
		t.Fatalf("bad status: want 400 got %d", c)
	}
	if c := setStatus(t, srv, 100, "inc_999999", "in_progress"); c != 404 {
		t.Fatalf("missing: want 404 got %d", c)
	}
	for _, st := range []string{"in_progress", "verifying", "done"} {
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, authedReq(t, "PATCH", "/api/incidents/"+inc.Id+"/status", `{"status":"`+st+`"}`, 100, "D"))
		var got incidentResp
		_ = json.Unmarshal(w.Body.Bytes(), &got)
		if w.Code != 200 || got.Status != st || got.Id != inc.Id {
			t.Fatalf("→ %s: %d %s", st, w.Code, w.Body)
		}
	}
	if c := setStatus(t, srv, 100, inc.Id, "done"); c != 422 {
		t.Fatalf("done → done: want 422 got %d", c)
	}
	// verifying notified both subscribers; done stamped resolved_at
	var n int
	if err := s.DB().QueryRowContext(context.Background(), `SELECT count(*) FROM outbox_message WHERE kind='incident_verifying'`).Scan(&n); err != nil || n != 2 {
		t.Fatalf("outbox verifying: n=%d err=%v", n, err)
	}
	var resolved bool
	if err := s.DB().QueryRowContext(context.Background(), `SELECT resolved_at IS NOT NULL FROM incident`).Scan(&resolved); err != nil || !resolved {
		t.Fatalf("resolved_at: %v %v", resolved, err)
	}
}

func TestUkQueueAndHouseStats(t *testing.T) {
	s := testStore(t)
	seedUK(t, s)
	srv := newTestServer(t, s)
	h10 := bindUser(t, srv, 1, "f-10")
	bindUser(t, srv, 3, "f-12")
	makeDispatcher(t, srv, s, 100)
	createIncident(t, srv, 1, `{"title":"A","description":"x","severity":"warning"}`)
	inc, _ := createIncident(t, srv, 1, `{"title":"B","description":"x","severity":"critical"}`)
	createIncident(t, srv, 3, `{"title":"C","description":"x","severity":"warning"}`)
	setStatus(t, srv, 100, inc.Id, "in_progress")

	var q struct {
		Items []struct {
			Id, HouseId, HouseAddress, Severity, Status, ReporterName string
			DueAt                                                     *string
			AffectedCount, ConfirmedCount                             int
		}
		Total int
	}
	if c := getJSON(t, srv, "/api/uk/queue?limit=2", 100, &q); c != 200 {
		t.Fatalf("queue: %d", c)
	}
	if q.Total != 3 || len(q.Items) != 2 || q.Items[0].Id != inc.Id || q.Items[0].Severity != "critical" {
		t.Fatalf("queue: %+v", q)
	}
	it := q.Items[0]
	if it.HouseId != h10 || it.HouseAddress != "f-10" || it.ReporterName != "U T." || it.DueAt == nil || it.AffectedCount != 1 || it.Status != "in_progress" {
		t.Fatalf("queue item: %+v", it)
	}
	if c := getJSON(t, srv, "/api/uk/queue", 1, nil); c != 403 {
		t.Fatalf("resident queue: want 403 got %d", c)
	}

	var st struct {
		ActiveIncidents, InProgress, ResolvedLast30Days int
		AvgResolutionHours                              *float64
		LastIncidentAt                                  *string
	}
	if c := getJSON(t, srv, "/api/houses/"+h10+"/stats", 1, &st); c != 200 {
		t.Fatalf("stats: %d", c)
	}
	if st.ActiveIncidents != 2 || st.InProgress != 1 || st.ResolvedLast30Days != 0 || st.AvgResolutionHours != nil || st.LastIncidentAt == nil {
		t.Fatalf("stats: %+v", st)
	}
	if c := getJSON(t, srv, "/api/houses/"+h10+"/stats", 3, nil); c != 404 {
		t.Fatalf("stats of another house: want 404 got %d", c)
	}
	if c := getJSON(t, srv, "/api/houses/"+h10+"/stats", 100, nil); c != 200 {
		t.Fatalf("dispatcher stats: want 200 got %d", c)
	}
	setStatus(t, srv, 100, inc.Id, "verifying")
	setStatus(t, srv, 100, inc.Id, "done")
	getJSON(t, srv, "/api/houses/"+h10+"/stats", 1, &st)
	if st.ActiveIncidents != 1 || st.ResolvedLast30Days != 1 || st.AvgResolutionHours == nil {
		t.Fatalf("stats after done: %+v", st)
	}
}

func TestEvents(t *testing.T) {
	s := testStore(t)
	seedUK(t, s)
	srv := newTestServer(t, s)
	h10 := bindUser(t, srv, 1, "f-10")
	bindUser(t, srv, 3, "f-12")
	makeDispatcher(t, srv, s, 100)

	body := `{"houseId":"` + h10 + `","reason":"Отключение ГВС","responsible":"УК-1","scheduledFrom":"2026-09-20T09:00:00Z","scheduledTo":"2026-09-20T13:00:00Z"}`
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/uk/events", body, 1, "U"))
	if w.Code != 403 {
		t.Fatalf("resident create: want 403 got %d", w.Code)
	}
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/uk/events", `{"houseId":"`+h10+`","reason":"","responsible":"x","scheduledFrom":"2026-09-20T09:00:00Z","scheduledTo":"2026-09-20T08:00:00Z"}`, 100, "D"))
	if w.Code != 400 {
		t.Fatalf("invalid: want 400 got %d %s", w.Code, w.Body)
	}
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/uk/events", body, 100, "D"))
	var ev struct {
		Id, HouseId, Status  string
		Entrance, ResolvedAt *string
	}
	_ = json.Unmarshal(w.Body.Bytes(), &ev)
	if w.Code != 201 || ev.HouseId != h10 || ev.Status != "planned" || ev.Entrance != nil || ev.ResolvedAt != nil {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"entrance":null`)) {
		t.Fatalf("nullable keys must be present: %s", w.Body)
	}

	var list struct{ Items []struct{ Id string } }
	if c := getJSON(t, srv, "/api/events?houseId="+h10, 1, &list); c != 200 || len(list.Items) != 1 || list.Items[0].Id != ev.Id {
		t.Fatalf("list: %d %+v", c, list)
	}
	if c := getJSON(t, srv, "/api/events?houseId="+h10, 3, nil); c != 404 {
		t.Fatalf("list of another house: want 404 got %d", c)
	}
	if c := getJSON(t, srv, "/api/events/"+ev.Id, 1, nil); c != 200 {
		t.Fatalf("get: %d", c)
	}
	if c := getJSON(t, srv, "/api/events/"+ev.Id, 3, nil); c != 404 {
		t.Fatalf("get from another house: want 404 got %d", c)
	}
}

// tiny valid 1x1 PNG
var pngBytes = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\rIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82")

func uploadPhoto(t *testing.T, srv http.Handler, incID string, field string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile(field, "p.png")
	_, _ = fw.Write(data)
	_ = mw.Close()
	r := authedReq(t, "POST", "/api/incidents/"+incID+"/photos", "x", 1, "U")
	r.Body = httptest.NewRequest("POST", "/", &buf).Body
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	return w
}

func TestIncidentPhotos(t *testing.T) {
	s := testStore(t)
	seedUK(t, s)
	srv := newTestServer(t, s)
	bindUser(t, srv, 1, "f-10")
	inc, _ := createIncident(t, srv, 1, `{"title":"A","description":"x","severity":"warning"}`)

	if w := uploadPhoto(t, srv, inc.Id, "photo", []byte("hello, not an image")); w.Code != 400 {
		t.Fatalf("text: want 400 got %d %s", w.Code, w.Body)
	}
	if w := uploadPhoto(t, srv, inc.Id, "file", pngBytes); w.Code != 400 {
		t.Fatalf("wrong field: want 400 got %d", w.Code)
	}
	if w := uploadPhoto(t, srv, "inc_999999", "photo", pngBytes); w.Code != 404 {
		t.Fatalf("missing incident: want 404 got %d", w.Code)
	}
	w := uploadPhoto(t, srv, inc.Id, "photo", pngBytes)
	var ph struct{ Id, Url string }
	_ = json.Unmarshal(w.Body.Bytes(), &ph)
	if w.Code != 201 || ph.Url != "/api/photos/"+ph.Id {
		t.Fatalf("upload: %d %s", w.Code, w.Body)
	}

	// public download, no Authorization header
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest("GET", ph.Url, nil))
	if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" || !bytes.Equal(w.Body.Bytes(), pngBytes) {
		t.Fatalf("download: %d %s", w.Code, w.Header().Get("Content-Type"))
	}
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest("GET", "/api/photos/ph_999", nil))
	if w.Code != 404 {
		t.Fatalf("missing photo: %d", w.Code)
	}

	var got struct{ Photos []struct{ Id, Url string } }
	if c := getJSON(t, srv, "/api/incidents/"+inc.Id, 1, &got); c != 200 || len(got.Photos) != 1 || got.Photos[0].Id != ph.Id {
		t.Fatalf("incident photos: %d %+v", c, got)
	}
}
