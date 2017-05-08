package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var fakeFSHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("fs")) })

func TestBasicAuth(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.Handler(BASICAUTH(fakeFSHandler))

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("basicauth handler wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestHideRootDotfiles(t *testing.T) {
	req, err := http.NewRequest("GET", "/.test", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.Handler(hideRootDotfiles(fakeFSHandler))

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("hideRootDotfiles handler wrong status code: got %v want %v", status, http.StatusForbidden)
	}
}
