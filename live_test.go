package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParsePatterns(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"*", []string{"*"}},
		{"*.html,*.css,*.js", []string{"*.html", "*.css", "*.js"}},
		{"*.html, *.css, *.js", []string{"*.html", "*.css", "*.js"}},
		{"", []string{"*"}},
		{"  ", []string{"*"}},
	}

	for _, tc := range tests {
		got := parsePatterns(tc.input)
		if len(got) != len(tc.expected) {
			t.Errorf("parsePatterns(%q) = %v, want %v", tc.input, got, tc.expected)
			continue
		}
		for i := range got {
			if got[i] != tc.expected[i] {
				t.Errorf("parsePatterns(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.expected[i])
			}
		}
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		patterns []string
		filename string
		expected bool
	}{
		{[]string{"*"}, "test.html", true},
		{[]string{"*"}, "test.css", true},
		{[]string{"*.html"}, "test.html", true},
		{[]string{"*.html"}, "test.css", false},
		{[]string{"*.html", "*.css"}, "test.css", true},
		{[]string{"*.html", "*.css"}, "test.js", false},
		{[]string{"*.html", "*.css", "*.js"}, "app.js", true},
	}

	for _, tc := range tests {
		lr := &LiveReload{patterns: tc.patterns}
		got := lr.matchesPattern(tc.filename)
		if got != tc.expected {
			t.Errorf("matchesPattern(%v, %q) = %v, want %v", tc.patterns, tc.filename, got, tc.expected)
		}
	}
}

func TestSSEEndpoint(t *testing.T) {
	lr, err := NewLiveReload("*")
	if err != nil {
		t.Fatal(err)
	}
	defer lr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/__livereload", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		lr.ServeHTTP(rec, req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	lr.Notify()
	time.Sleep(50 * time.Millisecond)

	// Cancel context and wait for handler to finish before reading buffer
	cancel()
	<-done

	result := rec.Body.String()
	if !strings.Contains(result, "data: reload") {
		t.Errorf("expected SSE response to contain 'data: reload', got %q", result)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected Content-Type 'text/event-stream', got %q", ct)
	}
}

func TestScriptInjection(t *testing.T) {
	lr, err := NewLiveReload("*")
	if err != nil {
		t.Fatal(err)
	}
	defer lr.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<!DOCTYPE html><html><body><h1>Test</h1></body></html>"))
	})

	wrapped := liveReloadMiddleware(handler, lr)

	req := httptest.NewRequest(http.MethodGet, "/test.html", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "EventSource") {
		t.Error("expected live reload script to be injected")
	}
	if !strings.Contains(body, "</body></html>") {
		t.Error("expected closing tags to be preserved")
	}
}

func TestScriptNotInjectedForNonHTML(t *testing.T) {
	lr, err := NewLiveReload("*")
	if err != nil {
		t.Fatal(err)
	}
	defer lr.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"test": true}`))
	})

	wrapped := liveReloadMiddleware(handler, lr)

	req := httptest.NewRequest(http.MethodGet, "/api/data.json", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "EventSource") {
		t.Error("script should not be injected into non-HTML responses")
	}
	if body != `{"test": true}` {
		t.Errorf("expected original body, got %q", body)
	}
}

func TestFileWatcher(t *testing.T) {
	tmpDir := t.TempDir()

	lr, err := NewLiveReload("*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer lr.Close()

	if err := lr.Watch(tmpDir); err != nil {
		t.Fatal(err)
	}
	lr.Start()

	notified := make(chan struct{}, 1)
	ch := make(chan struct{}, 1)
	lr.mu.Lock()
	lr.clients[ch] = true
	lr.mu.Unlock()

	go func() {
		<-ch
		notified <- struct{}{}
	}()

	testFile := filepath.Join(tmpDir, "test.html")
	if err := os.WriteFile(testFile, []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-notified:
	case <-time.After(500 * time.Millisecond):
		t.Error("expected notification for file change, timed out")
	}
}

func TestLiveReloadScriptContent(t *testing.T) {
	if !bytes.Contains(liveReloadScript, []byte("EventSource")) {
		t.Error("script should use EventSource for SSE")
	}
	if !bytes.Contains(liveReloadScript, []byte("/__livereload")) {
		t.Error("script should connect to /__livereload endpoint")
	}
	if !bytes.Contains(liveReloadScript, []byte("location.reload()")) {
		t.Error("script should call location.reload() on message")
	}
}
