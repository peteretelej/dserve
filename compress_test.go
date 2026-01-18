package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShouldCompress(t *testing.T) {
	tests := []struct {
		contentType string
		want        bool
	}{
		{"text/html", true},
		{"text/html; charset=utf-8", true},
		{"text/css", true},
		{"text/plain", true},
		{"text/xml", true},
		{"application/javascript", true},
		{"application/json", true},
		{"application/xml", true},
		{"application/xhtml+xml", true},
		{"image/svg+xml", true},
		{"image/png", false},
		{"image/jpeg", false},
		{"image/gif", false},
		{"video/mp4", false},
		{"audio/mpeg", false},
		{"application/octet-stream", false},
		{"application/pdf", false},
		{"font/woff2", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			if got := shouldCompress(tt.contentType); got != tt.want {
				t.Errorf("shouldCompress(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestGzipMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body>Hello World</body></html>"))
	})

	wrapped := gzipMiddleware(handler)

	t.Run("compresses when client accepts gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip, deflate")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") != "gzip" {
			t.Error("expected Content-Encoding: gzip")
		}
		if rec.Header().Get("Vary") != "Accept-Encoding" {
			t.Error("expected Vary: Accept-Encoding")
		}

		gr, err := gzip.NewReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer gr.Close()

		body, err := io.ReadAll(gr)
		if err != nil {
			t.Fatalf("failed to read gzip body: %v", err)
		}

		if !strings.Contains(string(body), "Hello World") {
			t.Error("decompressed body should contain 'Hello World'")
		}
	})

	t.Run("does not compress when client does not accept gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("should not set Content-Encoding: gzip when client doesn't accept it")
		}

		body := rec.Body.String()
		if !strings.Contains(body, "Hello World") {
			t.Error("body should contain 'Hello World'")
		}
	})
}

func TestGzipMiddlewareSkipsNonCompressible(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte{0x89, 0x50, 0x4E, 0x47}) // PNG magic bytes
	})

	wrapped := gzipMiddleware(handler)

	req := httptest.NewRequest("GET", "/image.png", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Error("should not compress image/png")
	}
}

func TestGzipResponseWriter(t *testing.T) {
	t.Run("handles multiple writes", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("first "))
			w.Write([]byte("second "))
			w.Write([]byte("third"))
		})

		wrapped := gzipMiddleware(handler)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		gr, err := gzip.NewReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer gr.Close()

		body, _ := io.ReadAll(gr)
		if string(body) != "first second third" {
			t.Errorf("expected 'first second third', got %q", string(body))
		}
	})
}
