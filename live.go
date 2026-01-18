package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var liveReloadScript = []byte(`<script>
(function(){
  var es = new EventSource('/__livereload');
  es.onmessage = function(e) { if(e.data==='reload') location.reload(); };
  es.onerror = function() { setTimeout(function(){ location.reload(); }, 1000); };
})();
</script>`)

type LiveReload struct {
	clients   map[chan struct{}]bool
	mu        sync.RWMutex
	watcher   *fsnotify.Watcher
	patterns  []string
	debouncer *time.Timer
	debounceMu sync.Mutex
}

func NewLiveReload(patterns string) (*LiveReload, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	lr := &LiveReload{
		clients:  make(map[chan struct{}]bool),
		watcher:  watcher,
		patterns: parsePatterns(patterns),
	}

	return lr, nil
}

func parsePatterns(s string) []string {
	var patterns []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			patterns = append(patterns, p)
		}
	}
	if len(patterns) == 0 {
		patterns = []string{"*"}
	}
	return patterns
}

func (lr *LiveReload) Watch(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return lr.watcher.Add(path)
		}
		return nil
	})
}

func (lr *LiveReload) Start() {
	go func() {
		for {
			select {
			case event, ok := <-lr.watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					if lr.matchesPattern(event.Name) {
						lr.notifyDebounced()
					}
				}
			case _, ok := <-lr.watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
}

func (lr *LiveReload) notifyDebounced() {
	lr.debounceMu.Lock()
	defer lr.debounceMu.Unlock()
	if lr.debouncer != nil {
		lr.debouncer.Stop()
	}
	lr.debouncer = time.AfterFunc(100*time.Millisecond, lr.Notify)
}

func (lr *LiveReload) matchesPattern(name string) bool {
	if len(lr.patterns) == 1 && lr.patterns[0] == "*" {
		return true
	}
	ext := filepath.Ext(name)
	base := filepath.Base(name)
	for _, p := range lr.patterns {
		if p == "*" {
			return true
		}
		if strings.HasPrefix(p, "*.") {
			if ext == "."+p[2:] {
				return true
			}
		} else if matched, _ := filepath.Match(p, base); matched {
			return true
		}
	}
	return false
}

func (lr *LiveReload) Notify() {
	lr.mu.RLock()
	defer lr.mu.RUnlock()
	for ch := range lr.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (lr *LiveReload) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan struct{}, 1)
	lr.mu.Lock()
	lr.clients[ch] = true
	lr.mu.Unlock()

	defer func() {
		lr.mu.Lock()
		delete(lr.clients, ch)
		lr.mu.Unlock()
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-ch:
			fmt.Fprintf(w, "data: reload\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (lr *LiveReload) Close() error {
	return lr.watcher.Close()
}

func liveReloadMiddleware(next http.Handler, lr *LiveReload) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		isHTML := strings.HasSuffix(path, ".html") ||
			strings.HasSuffix(path, ".htm") ||
			path == "/" ||
			(strings.HasSuffix(path, "/") && !strings.Contains(filepath.Base(path), "."))

		if !isHTML {
			next.ServeHTTP(w, r)
			return
		}

		rec := &liveResponseRecorder{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(rec, r)

		body := rec.body.Bytes()
		contentType := rec.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/html") {
			rec.Header().Del("Content-Length")
			w.WriteHeader(rec.statusCode)
			w.Write(body)
			return
		}

		if idx := bytes.LastIndex(body, []byte("</body>")); idx != -1 {
			body = append(body[:idx], append(liveReloadScript, body[idx:]...)...)
		} else if idx := bytes.LastIndex(body, []byte("</html>")); idx != -1 {
			body = append(body[:idx], append(liveReloadScript, body[idx:]...)...)
		} else {
			body = append(body, liveReloadScript...)
		}

		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteHeader(rec.statusCode)
		w.Write(body)
	})
}

type liveResponseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	wroteHeader bool
}

func (r *liveResponseRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.statusCode = code
		r.wroteHeader = true
	}
}

func (r *liveResponseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.wroteHeader = true
	}
	return r.body.Write(b)
}

func (r *liveResponseRecorder) ReadFrom(src io.Reader) (int64, error) {
	return io.Copy(r.body, src)
}
