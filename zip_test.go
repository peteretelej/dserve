package main

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestZipHandler(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content2"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("hidden"), 0644)

	handler := zipHandler(tmpDir)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		checkZip   bool
	}{
		{
			name:       "download root",
			path:       "",
			wantStatus: http.StatusOK,
			checkZip:   true,
		},
		{
			name:       "download subdir",
			path:       "subdir",
			wantStatus: http.StatusOK,
			checkZip:   true,
		},
		{
			name:       "not found",
			path:       "nonexistent",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "not a directory",
			path:       "file1.txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "path traversal blocked",
			path:       "../../../etc",
			wantStatus: http.StatusNotFound, // Path normalization converts "../../../etc" to "etc" which doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/__zip?path="+tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.checkZip && rec.Code == http.StatusOK {
				if ct := rec.Header().Get("Content-Type"); ct != "application/zip" {
					t.Errorf("got Content-Type %q, want application/zip", ct)
				}

				zr, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
				if err != nil {
					t.Fatalf("failed to open zip: %v", err)
				}

				if len(zr.File) == 0 {
					t.Error("zip is empty")
				}
			}
		})
	}
}

func TestZipDotfileExclusion(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte("visible"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("hidden"), 0644)
	hiddenDir := filepath.Join(tmpDir, ".hiddendir")
	os.Mkdir(hiddenDir, 0755)
	os.WriteFile(filepath.Join(hiddenDir, "secret.txt"), []byte("secret"), 0644)

	handler := zipHandler(tmpDir)

	req := httptest.NewRequest("GET", "/__zip", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}

	zr, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}

	for _, f := range zr.File {
		if f.Name == ".hidden" || f.Name == ".hiddendir/secret.txt" {
			t.Errorf("dotfile %q should be excluded from zip", f.Name)
		}
	}

	found := false
	for _, f := range zr.File {
		if f.Name == "visible.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("visible.txt should be in zip")
	}
}

func TestZipDirectoryContents(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "folder")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "root.txt"), []byte("root content"), 0644)
	os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested content"), 0644)

	handler := zipHandler(tmpDir)

	req := httptest.NewRequest("GET", "/__zip", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	zr, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}

	expected := map[string]string{
		"root.txt":        "root content",
		"folder/nested.txt": "nested content",
	}

	for _, f := range zr.File {
		want, ok := expected[f.Name]
		if !ok {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Errorf("failed to open %s: %v", f.Name, err)
			continue
		}
		content, _ := io.ReadAll(rc)
		rc.Close()
		if string(content) != want {
			t.Errorf("%s: got %q, want %q", f.Name, content, want)
		}
		delete(expected, f.Name)
	}

	for name := range expected {
		t.Errorf("missing file in zip: %s", name)
	}
}
