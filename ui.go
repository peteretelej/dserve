package main

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

//go:embed web/ui.html
var uiHTML string

type fileInfo struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	IsDir    bool      `json:"isDir"`
}

func uiHandler(rootDir string, uploadEnabled, zipEnabled bool) http.Handler {
	fileServer := http.FileServer(http.Dir(rootDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path
		if urlPath == "" {
			urlPath = "/"
		}
		// Normalize URL path to prevent directory traversal
		cleanURLPath := path.Clean("/" + urlPath)
		relPath := strings.TrimPrefix(cleanURLPath, "/")
		fullPath := filepath.Join(rootDir, filepath.FromSlash(relPath))

		info, err := os.Stat(fullPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		if !strings.HasSuffix(r.URL.Path, "/") && r.URL.Path != "" {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
			return
		}

		entries, err := os.ReadDir(fullPath)
		if err != nil {
			http.Error(w, "cannot read directory", http.StatusInternalServerError)
			return
		}

		isRoot := relPath == "" || relPath == "."
		var files []fileInfo
		for _, e := range entries {
			if isRoot && strings.HasPrefix(e.Name(), ".") {
				continue
			}
			fi, err := e.Info()
			if err != nil {
				continue
			}
			files = append(files, fileInfo{
				Name:     e.Name(),
				Size:     fi.Size(),
				Modified: fi.ModTime(),
				IsDir:    e.IsDir(),
			})
		}

		if strings.Contains(r.Header.Get("Accept"), "application/json") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(files)
			return
		}

		displayPath := urlPath
		if !strings.HasSuffix(displayPath, "/") {
			displayPath += "/"
		}
		if displayPath == "./" {
			displayPath = "/"
		}

		filesJSON, _ := json.Marshal(files)
		pathJSON, _ := json.Marshal(displayPath)
		dataScript := `<script>window.DSERVE={files:` + string(filesJSON) +
			`,path:` + string(pathJSON) +
			`,uploadEnabled:` + boolStr(uploadEnabled) +
			`,zipEnabled:` + boolStr(zipEnabled) + `};</script>`

		html := strings.Replace(uiHTML, "<!-- DSERVE_DATA_PLACEHOLDER -->", dataScript, 1)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
