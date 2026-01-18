package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var fakeFSHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "fs") })

func TestAuthInit(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		wantNil bool
	}{
		{"", false, true},
		{"admin:password123", false, false},
		{"user:pass", false, false},
		{"ab:c", true, true},       // username too short
		{"abc:", true, true},       // no password
		{":password", true, true},  // no username
		{"nopassword", true, true}, // no colon
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			creds = nil
			err := authInit(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("authInit(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantNil && creds != nil {
				t.Errorf("authInit(%q) creds should be nil", tt.input)
			}
			if !tt.wantNil && creds == nil {
				t.Errorf("authInit(%q) creds should not be nil", tt.input)
			}
		})
	}
}

func TestAuthInitParsesCorrectly(t *testing.T) {
	creds = nil
	err := authInit("myuser:mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.Username != "myuser" {
		t.Errorf("expected username 'myuser', got %q", creds.Username)
	}
	if creds.Password != "mypassword" {
		t.Errorf("expected password 'mypassword', got %q", creds.Password)
	}
}

func TestValidBasicAuth(t *testing.T) {
	creds = &AuthCreds{
		Username: "test",
		Password: "want12345",
	}
	defer func() { creds = nil }()

	tests := []struct {
		username string
		password string
		valid    bool
	}{
		{"tester", "abc", false},
		{"test", "want12345", true},
		{"test", "wrongpass", false},
		{"wronguser", "want12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			auth := []byte(tt.username + ":" + tt.password)
			req.Header.Set("Authorization",
				fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(auth)))

			valid := validBasicAuth(req)
			if tt.valid != valid {
				t.Errorf("expected valid %t, got %t", tt.valid, valid)
			}
		})
	}
}

func TestValidBasicAuthNoCreds(t *testing.T) {
	creds = nil
	req, _ := http.NewRequest("GET", "/", nil)
	if validBasicAuth(req) {
		t.Error("validBasicAuth should return false when creds is nil")
	}
}

func TestValidBasicAuthNoHeader(t *testing.T) {
	creds = &AuthCreds{Username: "test", Password: "pass"}
	defer func() { creds = nil }()

	req, _ := http.NewRequest("GET", "/", nil)
	if validBasicAuth(req) {
		t.Error("validBasicAuth should return false when no auth header")
	}
}

func TestBASICAUTH(t *testing.T) {
	creds = &AuthCreds{Username: "user", Password: "pass"}
	defer func() { creds = nil }()

	handler := BASICAUTH(fakeFSHandler)

	t.Run("rejects unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
		if rec.Header().Get("WWW-Authenticate") == "" {
			t.Error("expected WWW-Authenticate header")
		}
	})

	t.Run("allows authenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		auth := base64.StdEncoding.EncodeToString([]byte("user:pass"))
		req.Header.Set("Authorization", "Basic "+auth)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Body.String() != "fs" {
			t.Errorf("expected body 'fs', got %q", rec.Body.String())
		}
	})
}

func TestHideRootDotfiles(t *testing.T) {
	handler := hideRootDotfiles(fakeFSHandler)

	tests := []struct {
		path       string
		wantStatus int
	}{
		{"/.hidden", http.StatusForbidden},
		{"/.git", http.StatusForbidden},
		{"/.env", http.StatusForbidden},
		{"/visible", http.StatusOK},
		{"/path/to/.hidden", http.StatusOK}, // only root dotfiles blocked
		{"/", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("path %q: expected %d, got %d", tt.path, tt.wantStatus, rec.Code)
			}
		})
	}
}
