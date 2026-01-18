package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var compressibleTypes = []string{
	"text/",
	"application/javascript",
	"application/json",
	"application/xml",
	"application/xhtml+xml",
	"image/svg+xml",
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gw          *gzip.Writer
	wroteHeader bool
	skipGzip    bool
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	ct := w.Header().Get("Content-Type")
	if ct == "" || !shouldCompress(ct) {
		w.skipGzip = true
	}

	if !w.skipGzip {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if w.skipGzip {
		return w.ResponseWriter.Write(b)
	}
	return w.gw.Write(b)
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip compression for Range requests (needed for video seeking, resumable downloads)
		if r.Header.Get("Range") != "" {
			next.ServeHTTP(w, r)
			return
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzipWriterPool.Get().(*gzip.Writer)
		gz.Reset(w)
		defer func() {
			gz.Close()
			gzipWriterPool.Put(gz)
		}()

		gzw := &gzipResponseWriter{ResponseWriter: w, gw: gz}
		next.ServeHTTP(gzw, r)
	})
}

func shouldCompress(contentType string) bool {
	ct := strings.ToLower(contentType)
	for _, prefix := range compressibleTypes {
		if strings.HasPrefix(ct, prefix) {
			return true
		}
	}
	return false
}
