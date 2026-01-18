package main

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func zipHandler(rootDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := r.URL.Query().Get("path")
		if reqPath == "" {
			reqPath = "/"
		}

		// Normalize URL path to prevent directory traversal
		cleanURLPath := path.Clean("/" + reqPath)
		relPath := strings.TrimPrefix(cleanURLPath, "/")
		fullPath := filepath.Join(rootDir, filepath.FromSlash(relPath))

		absRoot, err := filepath.Abs(rootDir)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Validate path is within root using filepath.Rel
		rel, err := filepath.Rel(absRoot, absPath)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		info, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !info.IsDir() {
			http.Error(w, "not a directory", http.StatusBadRequest)
			return
		}

		dirName := filepath.Base(absPath)
		if dirName == "." || dirName == "/" {
			dirName = "download"
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="`+dirName+`.zip"`)

		if err := zipDirectory(w, absPath, absRoot); err != nil {
			// Headers already sent, can't change status. Log for debugging.
			log.Printf("zip write error for %s: %v", absPath, err)
			return
		}
	})
}

func zipDirectory(w io.Writer, dir string, root string) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			relToRoot, _ := filepath.Rel(root, path)
			base := filepath.Base(path)
			if base != "." && strings.HasPrefix(base, ".") && filepath.Dir(relToRoot) == "." {
				return filepath.SkipDir
			}
			return nil
		}

		relToRoot, _ := filepath.Rel(root, path)
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") && filepath.Dir(relToRoot) == "." {
			return nil
		}

		relToDir, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return nil
		}
		header.Name = filepath.ToSlash(relToDir)
		header.Method = zip.Deflate

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		if _, err := io.Copy(writer, file); err != nil {
			return err
		}
		return nil
	})
}
