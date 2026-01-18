package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUIHandler(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "image.png"), []byte("fake png"), 0644)
	_ = os.Mkdir(filepath.Join(dir, "subdir"), 0755)
	_ = os.WriteFile(filepath.Join(dir, "subdir", "nested.txt"), []byte("nested"), 0644)

	handler := uiHandler(dir, false, false, false)

	t.Run("serves HTML for directory", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/html") {
			t.Errorf("expected Content-Type to contain text/html, got %s", contentType)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "window.DSERVE") {
			t.Error("response should contain window.DSERVE data")
		}
		if !strings.Contains(body, "file.txt") {
			t.Error("response should contain file.txt")
		}
		if !strings.Contains(body, "subdir") {
			t.Error("response should contain subdir")
		}
	})

	t.Run("serves JSON with Accept header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", contentType)
		}

		var files []fileInfo
		if err := json.NewDecoder(rec.Body).Decode(&files); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		if len(files) != 3 {
			t.Errorf("expected 3 files, got %d", len(files))
		}
	})

	t.Run("serves file directly", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/file.txt", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if rec.Body.String() != "hello" {
			t.Errorf("expected 'hello', got %q", rec.Body.String())
		}
	})

	t.Run("redirects directory without trailing slash", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/subdir", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMovedPermanently {
			t.Errorf("expected status 301, got %d", rec.Code)
		}

		location := rec.Header().Get("Location")
		if location != "/subdir/" {
			t.Errorf("expected redirect to /subdir/, got %s", location)
		}
	})

	t.Run("serves subdirectory", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/subdir/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "nested.txt") {
			t.Error("response should contain nested.txt")
		}
	})

	t.Run("returns 404 for missing file", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent.txt", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})
}

func TestUIHandlerWithFeatures(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "test.txt"), []byte("test"), 0644)

	t.Run("includes upload flag in data", func(t *testing.T) {
		handler := uiHandler(dir, true, false, false)
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		body := rec.Body.String()
		if !strings.Contains(body, "uploadEnabled:true") {
			t.Error("response should contain uploadEnabled:true")
		}
		if !strings.Contains(body, "zipEnabled:false") {
			t.Error("response should contain zipEnabled:false")
		}
	})

	t.Run("includes zip flag in data", func(t *testing.T) {
		handler := uiHandler(dir, false, true, false)
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		body := rec.Body.String()
		if !strings.Contains(body, "uploadEnabled:false") {
			t.Error("response should contain uploadEnabled:false")
		}
		if !strings.Contains(body, "zipEnabled:true") {
			t.Error("response should contain zipEnabled:true")
		}
	})
}

func TestUIHandlerHidesDotfiles(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, ".hidden"), []byte("secret"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("public"), 0644)

	handler := uiHandler(dir, false, false, false)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var files []fileInfo
	_ = json.NewDecoder(rec.Body).Decode(&files)

	for _, f := range files {
		if f.Name == ".hidden" {
			t.Error("dotfile should be hidden in root directory listing")
		}
	}

	if len(files) != 1 || files[0].Name != "visible.txt" {
		t.Errorf("expected only visible.txt, got %v", files)
	}
}

func TestBoolStr(t *testing.T) {
	if boolStr(true) != "true" {
		t.Error("boolStr(true) should return 'true'")
	}
	if boolStr(false) != "false" {
		t.Error("boolStr(false) should return 'false'")
	}
}
