package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSpaMiddleware(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("SPA Index"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "app.html"), []byte("Custom App"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "style.css"), []byte("body{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "main.js"), []byte("console.log('hi')"), 0644); err != nil {
		t.Fatal(err)
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	fileServer := http.FileServer(http.Dir("."))

	t.Run("serves existing file", func(t *testing.T) {
		wrapped := spaMiddleware(fileServer, "index.html")
		req := httptest.NewRequest("GET", "/style.css", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Body.String() != "body{}" {
			t.Errorf("expected 'body{}', got %q", rec.Body.String())
		}
	})

	t.Run("serves file in subdirectory", func(t *testing.T) {
		wrapped := spaMiddleware(fileServer, "index.html")
		req := httptest.NewRequest("GET", "/assets/main.js", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Body.String() != "console.log('hi')" {
			t.Errorf("expected js content, got %q", rec.Body.String())
		}
	})

	t.Run("serves index.html for missing path", func(t *testing.T) {
		wrapped := spaMiddleware(fileServer, "index.html")
		req := httptest.NewRequest("GET", "/app/users/123", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Body.String() != "SPA Index" {
			t.Errorf("expected 'SPA Index', got %q", rec.Body.String())
		}
	})

	t.Run("serves custom fallback file", func(t *testing.T) {
		wrapped := spaMiddleware(fileServer, "app.html")
		req := httptest.NewRequest("GET", "/some/route", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Body.String() != "Custom App" {
			t.Errorf("expected 'Custom App', got %q", rec.Body.String())
		}
	})

	t.Run("serves root index.html", func(t *testing.T) {
		wrapped := spaMiddleware(fileServer, "index.html")
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}
