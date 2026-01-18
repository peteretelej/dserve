package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRangeRequests(t *testing.T) {
	dir := t.TempDir()
	content := "0123456789abcdefghij" // 20 bytes
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	fs := http.FileServer(http.Dir(dir))

	t.Run("returns 206 for Range request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test.txt", nil)
		req.Header.Set("Range", "bytes=0-9")
		rec := httptest.NewRecorder()

		fs.ServeHTTP(rec, req)

		if rec.Code != http.StatusPartialContent {
			t.Errorf("expected status 206, got %d", rec.Code)
		}
	})

	t.Run("returns correct Content-Range header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test.txt", nil)
		req.Header.Set("Range", "bytes=0-9")
		rec := httptest.NewRecorder()

		fs.ServeHTTP(rec, req)

		contentRange := rec.Header().Get("Content-Range")
		expected := "bytes 0-9/20"
		if contentRange != expected {
			t.Errorf("expected Content-Range %q, got %q", expected, contentRange)
		}
	})

	t.Run("returns correct partial content", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test.txt", nil)
		req.Header.Set("Range", "bytes=0-9")
		rec := httptest.NewRecorder()

		fs.ServeHTTP(rec, req)

		body := rec.Body.String()
		expected := "0123456789"
		if body != expected {
			t.Errorf("expected body %q, got %q", expected, body)
		}
	})

	t.Run("returns Accept-Ranges header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test.txt", nil)
		rec := httptest.NewRecorder()

		fs.ServeHTTP(rec, req)

		acceptRanges := rec.Header().Get("Accept-Ranges")
		if acceptRanges != "bytes" {
			t.Errorf("expected Accept-Ranges: bytes, got %q", acceptRanges)
		}
	})

	t.Run("handles middle range", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test.txt", nil)
		req.Header.Set("Range", "bytes=5-14")
		rec := httptest.NewRecorder()

		fs.ServeHTTP(rec, req)

		if rec.Code != http.StatusPartialContent {
			t.Errorf("expected status 206, got %d", rec.Code)
		}

		body := rec.Body.String()
		expected := "56789abcde"
		if body != expected {
			t.Errorf("expected body %q, got %q", expected, body)
		}
	})

	t.Run("handles suffix range", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test.txt", nil)
		req.Header.Set("Range", "bytes=-5")
		rec := httptest.NewRecorder()

		fs.ServeHTTP(rec, req)

		if rec.Code != http.StatusPartialContent {
			t.Errorf("expected status 206, got %d", rec.Code)
		}

		body := rec.Body.String()
		expected := "fghij" // last 5 bytes
		if body != expected {
			t.Errorf("expected body %q, got %q", expected, body)
		}
	})
}

func TestGzipMiddlewareSkipsRangeRequests(t *testing.T) {
	dir := t.TempDir()
	content := "0123456789abcdefghij" // 20 bytes
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fs := http.FileServer(http.Dir(dir))
	wrapped := gzipMiddleware(fs)

	t.Run("skips compression for Range request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test.txt", nil)
		req.Header.Set("Range", "bytes=0-9")
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Should NOT be gzip encoded
		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("should not compress Range requests")
		}

		// Should return 206 Partial Content
		if rec.Code != http.StatusPartialContent {
			t.Errorf("expected status 206, got %d", rec.Code)
		}

		// Body should be uncompressed partial content
		body := rec.Body.String()
		expected := "0123456789"
		if body != expected {
			t.Errorf("expected body %q, got %q", expected, body)
		}
	})

	t.Run("still compresses normal requests", func(t *testing.T) {
		// Create an HTML file that should be compressed
		htmlFile := filepath.Join(dir, "test.html")
		if err := os.WriteFile(htmlFile, []byte("<html><body>Hello World</body></html>"), 0644); err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("GET", "/test.html", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Should be gzip encoded for normal requests
		if rec.Header().Get("Content-Encoding") != "gzip" {
			t.Error("should compress normal HTML requests")
		}
	})
}
