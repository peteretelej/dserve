package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var fakeFSHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "fs") })

func TestValidBasicAuth(t *testing.T) {
	creds = &AuthCreds{
		Username: "test",
		Password: "want12345",
	}
	var tests = []struct {
		username string
		password string
		valid    bool
	}{
		{"tester", "abc", false},
		{"test", "want12345", true},
	}
	for _, test := range tests {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Errorf("failed to create request: %v", err)
			continue
		}
		auth := []byte(test.username + ":" + test.password)
		req.Header.Set("Authorization",
			fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(auth)))

		valid := validBasicAuth(req)
		if test.valid != valid {
			t.Errorf("expected valid %t, got %t", test.valid, valid)
		}
	}

}

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
