package http

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func uploadStaged(t *testing.T, srv http.Handler, maxID int64, data []byte) (int, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("photo", "p")
	_, _ = fw.Write(data)
	_ = mw.Close()
	r := authedReq(t, "POST", "/api/incidents/photos", "x", maxID, "U")
	r.Body = httptest.NewRequest("POST", "/", &buf).Body
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	var p struct{ Url string }
	_ = json.Unmarshal(w.Body.Bytes(), &p)
	return w.Code, p.Url
}

func TestStagedPhotos(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	bindUser(t, srv, 1, "f-10")

	if code, url := uploadStaged(t, srv, 1, pngBytes); code != 201 || url == "" {
		t.Fatalf("png: %d", code)
	}

	heic, err := os.ReadFile("testdata/sample.heic")
	if err != nil {
		t.Fatal(err)
	}
	code, url := uploadStaged(t, srv, 1, heic)
	if code != 201 {
		t.Fatalf("heic: %d", code)
	}
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest("GET", url, nil))
	if w.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("heic must be served as jpeg: %q", w.Header().Get("Content-Type"))
	}
	img, _, err := image.Decode(w.Body)
	if err != nil || img.Bounds().Dx() != 64 || img.Bounds().Dy() != 32 {
		t.Fatalf("decoded jpeg: %v", err)
	}

	junk := make([]byte, 10<<20)
	copy(junk, "\x00\x00\x00\x18ftypheic")
	for i := 16; i < len(junk); i++ {
		junk[i] = byte(i * 7919)
	}
	if code, _ := uploadStaged(t, srv, 1, junk); code != 400 {
		t.Fatalf("10 MB garbage heic: %d", code)
	}
	if code, _ := uploadStaged(t, srv, 1, heic[:40]); code != 400 {
		t.Fatalf("truncated heic: %d", code)
	}
	huge, err := os.ReadFile("testdata/huge.heic")
	if err != nil {
		t.Fatal(err)
	}
	if code, _ := uploadStaged(t, srv, 1, huge); code != 400 {
		t.Fatalf("heic over 50 MP must be rejected before decoding: %d", code)
	}
	if code, _ := uploadStaged(t, srv, 1, []byte("GIF89a......")); code != 400 {
		t.Fatalf("gif: %d", code)
	}
}

func TestStagedPhotosRateLimit(t *testing.T) {
	s := testStore(t)
	srv := newTestServer(t, s)
	bindUser(t, srv, 1, "f-10")
	p := pngBytes
	for i := 0; i < 20; i++ {
		if code, _ := uploadStaged(t, srv, 1, p); code != 201 {
			t.Fatalf("upload %d: %d", i, code)
		}
	}
	if code, _ := uploadStaged(t, srv, 1, p); code != 429 {
		t.Fatalf("21st: %d", code)
	}
	garbageHEIC := append([]byte("\x00\x00\x00\x18ftypheic"), bytes.Repeat([]byte{0xAB}, 1<<20)...)
	if code, _ := uploadStaged(t, srv, 1, garbageHEIC); code != 429 {
		t.Fatalf("rate limit must be checked before decoding: %d", code)
	}
}
