package handler

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/assets"
)

func TestRegisterWebCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	web := testWeb(t)
	r := gin.New()
	RegisterWeb(r, web)

	html := httptest.NewRecorder()
	r.ServeHTTP(html, httptest.NewRequest(http.MethodGet, "/", nil))
	if html.Code != http.StatusOK {
		t.Fatalf("index status = %d", html.Code)
	}
	if got := html.Header().Get("Cache-Control"); got != cacheControlHTML {
		t.Fatalf("index Cache-Control = %q", got)
	}
	if !bytes.Contains(html.Body.Bytes(), []byte("index-abc.js")) {
		t.Fatalf("index body = %q", html.Body.String())
	}

	js := httptest.NewRecorder()
	r.ServeHTTP(js, httptest.NewRequest(http.MethodGet, "/assets/index-abc.js", nil))
	if js.Code != http.StatusOK {
		t.Fatalf("js status = %d", js.Code)
	}
	if got := js.Header().Get("Cache-Control"); got != cacheControlHashed {
		t.Fatalf("js Cache-Control = %q", got)
	}
	if js.Body.String() != "console.log(1);\n" {
		t.Fatalf("js body = %q", js.Body.String())
	}
}

func TestRegisterWebGzipAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	web := testWeb(t)
	r := gin.New()
	RegisterWeb(r, web)

	req := httptest.NewRequest(http.MethodGet, "/assets/index-abc.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q", rec.Header().Get("Content-Encoding"))
	}
	if rec.Header().Get("Cache-Control") != cacheControlHashed {
		t.Fatalf("Cache-Control = %q", rec.Header().Get("Cache-Control"))
	}
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer gr.Close()
	body, err := io.ReadAll(gr)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "console.log(1);\n" {
		t.Fatalf("gunzip body = %q", body)
	}
}

func TestRegisterWebGzipIndexKeepsNoCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	web := testWeb(t)
	r := gin.New()
	RegisterWeb(r, web)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get("Cache-Control") != cacheControlHTML {
		t.Fatalf("Cache-Control = %q", rec.Header().Get("Cache-Control"))
	}
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q", rec.Header().Get("Content-Encoding"))
	}
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer gr.Close()
	body, err := io.ReadAll(gr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("index-abc.js")) {
		t.Fatalf("gunzip body = %q", body)
	}
}

func TestRegisterWebSkipsGzipWithoutAccept(t *testing.T) {
	gin.SetMode(gin.TestMode)
	web := testWeb(t)
	r := gin.New()
	RegisterWeb(r, web)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/assets/index-abc.js", nil))
	if rec.Header().Get("Content-Encoding") != "" {
		t.Fatalf("unexpected Content-Encoding %q", rec.Header().Get("Content-Encoding"))
	}
	if rec.Body.String() != "console.log(1);\n" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRegisterWebHeadAssetsSkipGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	web := testWeb(t)
	r := gin.New()
	RegisterWeb(r, web)

	req := httptest.NewRequest(http.MethodHead, "/assets/index-abc.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get("Cache-Control") != cacheControlHashed {
		t.Fatalf("Cache-Control = %q", rec.Header().Get("Cache-Control"))
	}
	if rec.Header().Get("Content-Encoding") != "" {
		t.Fatalf("HEAD should not gzip, got %q", rec.Header().Get("Content-Encoding"))
	}
}

func testWeb(t *testing.T) *assets.Web {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<script src="/assets/index-abc.js"></script>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "index-abc.js"), []byte("console.log(1);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return assets.FromDir(dir)
}
