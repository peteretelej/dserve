package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	dir       = flag.String("dir", "./", "directory to serve")
	port      = flag.Int("port", 9011, "port to serve on")
	local     = flag.Bool("local", false, "serve on localhost only")
	basicauth = flag.String("basicauth", "", "basic auth credentials (user:pass)")
	timeout   = flag.Duration("timeout", time.Minute*3, "server timeout")

	tlsEnabled = flag.Bool("tls", false, "enable HTTPS")
	certFile   = flag.String("cert", "", "TLS certificate file")
	keyFile    = flag.String("key", "", "TLS key file")

	compress = flag.Bool("compress", false, "enable gzip compression")
	spa      = flag.String("spa", "", "enable SPA mode with fallback file (default: index.html if flag present)")
	live     = flag.String("live", "", "enable live reload with watch pattern (default: * if flag present)")
	upload   = flag.Bool("upload", false, "enable file uploads")
	uploadDir = flag.String("upload-dir", "", "upload destination directory")
	maxSize  = flag.String("max-size", "100MB", "maximum upload size")
	zipDl    = flag.Bool("zip", false, "enable directory download as zip")
	webUI    = flag.Bool("webui", false, "enable web UI for directory listing")
	dotfiles = flag.Bool("dotfiles", false, "show and allow access to dotfiles (use with caution)")
)

func main() {
	flag.Parse()
	log.SetPrefix("dserve: ")

	if err := os.Chdir(*dir); err != nil {
		log.Fatal(err)
	}

	var addr string
	if *local {
		addr = "localhost"
	}

	if err := authInit(*basicauth); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg := &Config{
		Addr:     fmt.Sprintf("%s:%d", addr, *port),
		Timeout:  *timeout,
		Compress: *compress,
		Zip:      *zipDl,
		WebUI:    *webUI,
		Dotfiles: *dotfiles,
	}

	if cfg.Dotfiles {
		log.Println("WARNING: dotfiles are visible and accessible - ensure no sensitive files are exposed")
	}

	if *tlsEnabled {
		cfg.TLS = &TLSConfig{Cert: *certFile, Key: *keyFile}
	}

	if isFlagSet("spa") {
		cfg.SPA = *spa
		if cfg.SPA == "" {
			cfg.SPA = "index.html"
		}
	}

	if isFlagSet("live") {
		pattern := *live
		if pattern == "" {
			pattern = "*"
		}
		lr, err := NewLiveReload(pattern)
		if err != nil {
			log.Fatalf("Failed to initialize live reload: %v", err)
		}
		if err := lr.Watch("."); err != nil {
			log.Fatalf("Failed to watch directory: %v", err)
		}
		lr.Start()
		defer lr.Close()
		cfg.LiveReload = lr
	}

	if *upload {
		maxBytes, err := parseSize(*maxSize)
		if err != nil {
			log.Fatalf("invalid max-size: %v", err)
		}
		dest := *uploadDir
		if dest == "" {
			dest = "."
		}
		cfg.Upload = &UploadConfig{Dir: dest, MaxBytes: maxBytes}
	}

	protocol := "http"
	if cfg.TLS != nil {
		protocol = "https"
	}
	displayAddr := cfg.Addr
	if displayAddr[0] == ':' {
		displayAddr = "localhost" + displayAddr
	}
	fmt.Printf("Serving %s at %s://%s\n", *dir, protocol, displayAddr)
	if cfg.LiveReload != nil {
		fmt.Printf("Live reload enabled, watching: %s\n", *live)
	}
	if cfg.Upload != nil {
		fmt.Printf("Uploads enabled (max: %s, dest: %s)\n", *maxSize, cfg.Upload.Dir)
	}
	if cfg.WebUI {
		fmt.Printf("Browse files: %s://%s/__browse/\n", protocol, displayAddr)
	}

	if err := Serve(cfg); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func Serve(cfg *Config) error {
	mux := http.NewServeMux()

	if cfg.LiveReload != nil {
		mux.Handle("/__livereload", cfg.LiveReload)
	}

	uploadEnabled := cfg.Upload != nil
	if uploadEnabled {
		mux.Handle("/__upload", uploadHandler(cfg.Upload.Dir, cfg.Upload.MaxBytes))
	}

	if cfg.Zip {
		mux.Handle("/__zip", zipHandler("."))
	}

	if cfg.WebUI {
		mux.Handle("/__browse/", http.StripPrefix("/__browse", uiHandler(".", uploadEnabled, cfg.Zip, cfg.Dotfiles)))
	}

	var fs http.Handler
	if cfg.Dotfiles {
		fs = http.FileServer(http.Dir("."))
	} else {
		fs = hideRootDotfiles(http.FileServer(dotfileHidingFS{http.Dir(".")}))
	}

	if creds != nil {
		fs = BASICAUTH(fs)
	}

	if cfg.Compress {
		fs = gzipMiddleware(fs)
	}

	if cfg.SPA != "" {
		fs = spaMiddleware(fs, cfg.SPA)
	}

	if cfg.LiveReload != nil {
		fs = liveReloadMiddleware(fs, cfg.LiveReload)
	}

	mux.Handle("/", fs)

	svr := &http.Server{
		Addr:           cfg.Addr,
		Handler:        mux,
		ReadTimeout:    cfg.Timeout,
		WriteTimeout:   cfg.Timeout * 2,
		IdleTimeout:    cfg.Timeout * 10,
		MaxHeaderBytes: 1 << 20,
	}

	if cfg.TLS != nil {
		cert, key := cfg.TLS.Cert, cfg.TLS.Key
		if cert == "" || key == "" {
			var err error
			cert, key, err = loadOrGenerateCert()
			if err != nil {
				return fmt.Errorf("TLS setup failed: %w", err)
			}
		}
		return svr.ListenAndServeTLS(cert, key)
	}
	return svr.ListenAndServe()
}

func BASICAUTH(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validBasicAuth(r) {
			w.Header().Set("WWW-Authenticate", `Basic realm="dserve Basic Authentication"`)
			http.Error(w, "Not Authorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func hideRootDotfiles(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/.") {
			http.Error(w, "access to dotfiles in root directory is forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// dotfileHidingFS wraps http.FileSystem to hide dotfiles from root directory listings
type dotfileHidingFS struct {
	fs http.FileSystem
}

func (dfs dotfileHidingFS) Open(name string) (http.File, error) {
	f, err := dfs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return dotfileHidingFile{f, name == "/" || name == ""}, nil
}

type dotfileHidingFile struct {
	http.File
	isRoot bool
}

func (f dotfileHidingFile) Readdir(n int) ([]os.FileInfo, error) {
	files, err := f.File.Readdir(n)
	if err != nil || !f.isRoot {
		return files, err
	}
	filtered := files[:0]
	for _, fi := range files {
		if !strings.HasPrefix(fi.Name(), ".") {
			filtered = append(filtered, fi)
		}
	}
	return filtered, nil
}

type AuthCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var creds *AuthCreds

func authInit(bAuth string) error {
	if bAuth == "" {
		return nil
	}
	i := strings.Index(bAuth, ":")
	if i < 3 || i >= len(bAuth)-1 {
		return errors.New("invalid basicauth flag value: value should be USERNAME:PASSWORD, e.g. dserve -basicauth admin:passw0rd")
	}
	creds = &AuthCreds{
		Username: bAuth[:i],
		Password: bAuth[i+1:],
	}
	return nil
}

func validBasicAuth(r *http.Request) bool {
	if creds == nil {
		return false
	}
	u, p, ok := r.BasicAuth()
	if !ok {
		return ok
	}
	return u == creds.Username && p == creds.Password
}
