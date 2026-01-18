package main

import (
	"net/http"
	"os"
	"path/filepath"
)

func spaMiddleware(next http.Handler, indexFile string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(".", filepath.Clean(r.URL.Path))

		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			indexPath := filepath.Join(path, indexFile)
			if _, err := os.Stat(indexPath); err == nil {
				next.ServeHTTP(w, r)
				return
			}
		}

		if err != nil && os.IsNotExist(err) {
			http.ServeFile(w, r, indexFile)
			return
		}

		next.ServeHTTP(w, r)
	})
}
