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

func TestHouseStats(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	h10 := bindUser(t, srv, 1, "f-10")
	bindUser(t, srv, 3, "f-12")
	createIncident(t, srv, 1, `{"title":"A","description":"x","severity":"warning"}`)
	inc, _ := createIncident(t, srv, 1, `{"title":"B","description":"x","severity":"critical"}`)
	createIncident(t, srv, 3, `{"title":"C","description":"x","severity":"warning"}`)
	incID, _ := parseID("inc_", inc.Id)
	if _, err := s.DB().ExecContext(context.Background(), `UPDATE incident SET status='in_progress' WHERE id=$1`, incID); err != nil {
		t.Fatal(err)
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

	if _, err := s.DB().ExecContext(context.Background(),
		`UPDATE incident SET status='done', resolved_at=now() WHERE id=$1`, incID); err != nil {
		t.Fatal(err)
	}
	getJSON(t, srv, "/api/houses/"+h10+"/stats", 1, &st)
	if st.ActiveIncidents != 1 || st.ResolvedLast30Days != 1 || st.AvgResolutionHours == nil {
		t.Fatalf("stats after done: %+v", st)
	}
}

var pngBytes = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\rIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82")

func uploadPhoto(t *testing.T, srv http.Handler, incID string, field string, data []byte) *httptest.ResponseRecorder {
	return uploadPhotoAs(t, srv, 1, incID, field, data)
}

func uploadPhotoAs(t *testing.T, srv http.Handler, maxID int64, incID string, field string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile(field, "p.png")
	_, _ = fw.Write(data)
	_ = mw.Close()
	r := authedReq(t, "POST", "/api/incidents/"+incID+"/photos", "x", maxID, "U")
	r.Body = httptest.NewRequest("POST", "/", &buf).Body
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	return w
}

func TestIncidentPhotos(t *testing.T) {
	s := testStore(t)
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

	bindUser(t, srv, 2, "f-10")
	if w := uploadPhotoAs(t, srv, 2, inc.Id, "photo", pngBytes); w.Code != 403 {
		t.Fatalf("stranger: want 403 got %d", w.Code)
	}
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, authedReq(t, "POST", "/api/incidents/"+inc.Id+"/join", "", 2, "U"))
	if w := uploadPhotoAs(t, srv, 2, inc.Id, "photo", pngBytes); w.Code != 201 {
		t.Fatalf("subscriber: want 201 got %d", w.Code)
	}

	for i := 0; i < 3; i++ {
		if w := uploadPhoto(t, srv, inc.Id, "photo", pngBytes); w.Code != 201 {
			t.Fatalf("photo %d: %d %s", i+3, w.Code, w.Body)
		}
	}
	if w := uploadPhoto(t, srv, inc.Id, "photo", pngBytes); w.Code != 422 {
		t.Fatalf("6th photo: want 422 got %d", w.Code)
	}

	inc2, _ := createIncident(t, srv, 1, `{"title":"B","description":"x","severity":"warning"}`)
	if _, err := s.DB().ExecContext(context.Background(),
		`INSERT INTO incident_photo (incident_id, user_id, content_type, data) SELECT $1, user_id, 'image/png', ''::bytea FROM incident_photo, generate_series(1,4) WHERE incident_id=$2 AND user_id=(SELECT id FROM app_user WHERE max_user_id=1)`,
		mustID(t, inc2.Id), mustID(t, inc.Id)); err != nil {
		t.Fatal(err)
	}
	if w := uploadPhoto(t, srv, inc2.Id, "photo", pngBytes); w.Code != 429 {
		t.Fatalf("21st photo in an hour: want 429 got %d %s", w.Code, w.Body)
	}
}

func mustID(t *testing.T, wire string) int64 {
	t.Helper()
	id, ok := parseID("inc_", wire)
	if !ok {
		t.Fatalf("bad id %q", wire)
	}
	return id
}
