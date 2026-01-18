package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type uploadResponse struct {
	Success  bool   `json:"success"`
	Filename string `json:"filename,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Error    string `json:"error,omitempty"`
}

func uploadHandler(destDir string, maxSize int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(uploadResponse{Error: "method not allowed"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxSize)

		if err := r.ParseMultipartForm(maxSize); err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			json.NewEncoder(w).Encode(uploadResponse{Error: "file too large"})
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(uploadResponse{Error: "no file provided"})
			return
		}
		defer file.Close()

		filename := safeFilename(header.Filename)
		if filename == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(uploadResponse{Error: "invalid filename"})
			return
		}

		subdir := safeFilename(r.FormValue("path"))
		targetDir := destDir
		if subdir != "" {
			targetDir = filepath.Join(destDir, subdir)
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(uploadResponse{Error: "failed to create directory"})
				return
			}
		}

		destPath := uniqueFilename(targetDir, filename)

		dst, err := os.Create(destPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(uploadResponse{Error: "failed to create file"})
			return
		}
		defer dst.Close()

		size, err := io.Copy(dst, file)
		if err != nil {
			os.Remove(destPath)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(uploadResponse{Error: "failed to save file"})
			return
		}

		json.NewEncoder(w).Encode(uploadResponse{
			Success:  true,
			Filename: filepath.Base(destPath),
			Size:     size,
		})
	})
}

var safeFilenameRe = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func safeFilename(name string) string {
	name = filepath.Base(name)
	name = safeFilenameRe.ReplaceAllString(name, "_")
	name = strings.TrimLeft(name, ".")
	return name
}

func uniqueFilename(dir, name string) string {
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	for i := 1; i < 1000; i++ {
		newName := fmt.Sprintf("%s_%d%s", base, i, ext)
		path = filepath.Join(dir, newName)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}
	return path
}

func parseSize(s string) (int64, error) {
	s = strings.ToUpper(strings.TrimSpace(s))

	suffixes := []struct {
		suffix string
		mult   int64
	}{
		{"GB", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
		{"KB", 1024},
		{"B", 1},
	}

	for _, entry := range suffixes {
		if strings.HasSuffix(s, entry.suffix) {
			numStr := strings.TrimSuffix(s, entry.suffix)
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				return 0, err
			}
			return num * entry.mult, nil
		}
	}

	return strconv.ParseInt(s, 10, 64)
}
