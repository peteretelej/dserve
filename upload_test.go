package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"100B", 100, false},
		{"100b", 100, false},
		{"1KB", 1024, false},
		{"1kb", 1024, false},
		{"10MB", 10 * 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"100", 100, false},
		{"  50MB  ", 50 * 1024 * 1024, false},
		{"abc", 0, true},
		{"MB", 0, true},
	}

	for _, tt := range tests {
		result, err := parseSize(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseSize(%q) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("parseSize(%q) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("parseSize(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestSafeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"file.txt", "file.txt"},
		{"my file.txt", "my_file.txt"},
		{"../../../etc/passwd", "passwd"},
		{"/absolute/path/file.txt", "file.txt"},
		{".hidden", "hidden"},
		{"...dots", "dots"},
		{"file<>|:*.txt", "file_____.txt"},
		{"normal-file_123.txt", "normal-file_123.txt"},
	}

	for _, tt := range tests {
		result := safeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("safeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestUniqueFilename(t *testing.T) {
	dir := t.TempDir()

	path := uniqueFilename(dir, "test.txt")
	if filepath.Base(path) != "test.txt" {
		t.Errorf("uniqueFilename should return test.txt when file doesn't exist, got %s", filepath.Base(path))
	}

	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("content"), 0644)
	path = uniqueFilename(dir, "test.txt")
	if filepath.Base(path) != "test_1.txt" {
		t.Errorf("uniqueFilename should return test_1.txt when test.txt exists, got %s", filepath.Base(path))
	}

	os.WriteFile(filepath.Join(dir, "test_1.txt"), []byte("content"), 0644)
	path = uniqueFilename(dir, "test.txt")
	if filepath.Base(path) != "test_2.txt" {
		t.Errorf("uniqueFilename should return test_2.txt when test.txt and test_1.txt exist, got %s", filepath.Base(path))
	}
}

func TestUploadHandler(t *testing.T) {
	dir := t.TempDir()
	handler := uploadHandler(dir, 1024*1024)

	t.Run("successful upload", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write([]byte("hello world"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/__upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var resp uploadResponse
		json.NewDecoder(rec.Body).Decode(&resp)
		if !resp.Success {
			t.Errorf("expected success=true, got error: %s", resp.Error)
		}
		if resp.Filename != "test.txt" {
			t.Errorf("expected filename=test.txt, got %s", resp.Filename)
		}
		if resp.Size != 11 {
			t.Errorf("expected size=11, got %d", resp.Size)
		}

		content, _ := os.ReadFile(filepath.Join(dir, "test.txt"))
		if string(content) != "hello world" {
			t.Errorf("file content mismatch: %s", string(content))
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/__upload", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", rec.Code)
		}
	})

	t.Run("no file provided", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/__upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}

		var resp uploadResponse
		json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Error != "no file provided" {
			t.Errorf("expected error='no file provided', got %s", resp.Error)
		}
	})

	t.Run("file too large", func(t *testing.T) {
		smallHandler := uploadHandler(dir, 10)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "large.txt")
		part.Write(bytes.Repeat([]byte("x"), 1000))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/__upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		smallHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected status 413, got %d", rec.Code)
		}
	})

	t.Run("upload with subdirectory", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "subfile.txt")
		part.Write([]byte("subdir content"))
		writer.WriteField("path", "subdir")
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/__upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		content, err := os.ReadFile(filepath.Join(dir, "subdir", "subfile.txt"))
		if err != nil {
			t.Errorf("file not created in subdir: %v", err)
		}
		if string(content) != "subdir content" {
			t.Errorf("file content mismatch: %s", string(content))
		}
	})

	t.Run("path traversal blocked", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "../../../etc/passwd")
		part.Write([]byte("evil content"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/__upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var resp uploadResponse
		json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Filename != "passwd" {
			t.Errorf("expected sanitized filename 'passwd', got %s", resp.Filename)
		}

		if _, err := os.Stat(filepath.Join(dir, "..", "..", "..", "etc", "passwd")); err == nil {
			t.Error("path traversal should be blocked")
		}
	})
}

func createMultipartRequest(t *testing.T, filename string, content []byte) (*http.Request, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(part, bytes.NewReader(content))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/__upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, writer.FormDataContentType()
}
