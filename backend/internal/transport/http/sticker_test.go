package http

import (
	"encoding/json"
	"encoding/xml"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSticker_PublicSVGAndStartParamBindsHouse(t *testing.T) {
	repo := testStore(t)
	srv := newTestServer(t, repo)
	houseID := bindUser(t, srv, 1, "f-10")

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest("GET", "/api/houses/"+houseID+"/sticker", nil))
	if w.Code != 200 || w.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("sticker: %d %s", w.Code, w.Header().Get("Content-Type"))
	}
	body := w.Body.String()
	if !strings.Contains(body, `<path d="M`) {
		t.Fatalf("no qr in sticker: %s", body)
	}
	dec := xml.NewDecoder(strings.NewReader(body))
	for {
		if _, err := dec.Token(); err != nil {
			if err.Error() != "EOF" {
				t.Fatalf("sticker is not valid xml: %v", err)
			}
			break
		}
	}

	w = httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest("GET", "/api/houses/h_999/sticker", nil))
	if w.Code != 404 {
		t.Fatalf("unknown house: %d", w.Code)
	}

	r := authedReq(t, "GET", "/api/me", "", 2, "N")
	r.Header.Set("Authorization", "tma "+signInitData(t, map[string]string{
		"user":        `{"id":2,"first_name":"N"}`,
		"start_param": houseID,
	}))
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	var me struct{ House *struct{ Id string } }
	_ = json.Unmarshal(w.Body.Bytes(), &me)
	if w.Code != 200 || me.House == nil || me.House.Id != houseID {
		t.Fatalf("start_param must bind house: %d %s", w.Code, w.Body)
	}
}
