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
	dir        = flag.String("dir", "./", "the directory to serve, defaults to current directory")
	port       = flag.Int("port", 9011, "the port to serve at, defaults 9011")
	local      = flag.Bool("local", false, "whether to serve on all address or on localhost, default all addresses")
	basicauth  = flag.String("basicauth", "", "basicauth creds, enables basic authentication")
	timeout    = flag.Duration("timeout", time.Minute*3, "http server read timeout, write timeout will be double this")
	tlsEnabled = flag.Bool("tls", false, "enable HTTPS")
	certFile   = flag.String("cert", "", "TLS certificate file")
	keyFile    = flag.String("key", "", "TLS key file")
	compress   = flag.Bool("compress", false, "enable gzip compression")
	spa        = flag.Bool("spa", false, "enable SPA mode (serve index.html for missing routes)")
	spaIndex   = flag.String("spa-index", "index.html", "SPA fallback file")
	live       = flag.Bool("live", false, "enable live reload on file changes")
	watch      = flag.String("watch", "*", "file patterns to watch for live reload (comma-separated)")
	upload     = flag.Bool("upload", false, "enable file uploads")
	uploadDir  = flag.String("upload-dir", "", "upload destination directory")
	maxSize    = flag.String("max-size", "100MB", "maximum upload size")
	zipDl      = flag.Bool("zip", false, "enable directory download as zip")
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

	var lr *LiveReload
	if *live {
		var err error
		lr, err = NewLiveReload(*watch)
		if err != nil {
			log.Fatalf("Failed to initialize live reload: %v", err)
		}
		if err := lr.Watch("."); err != nil {
			log.Fatalf("Failed to watch directory: %v", err)
		}
		lr.Start()
		defer lr.Close()
	}

	listenAddr := fmt.Sprintf("%s:%d", addr, *port)
	protocol := "http"
	if *tlsEnabled {
		protocol = "https"
	}
	fmt.Printf("Launching dserve %s server %s on %s\n", protocol, *dir, listenAddr)
	if *live {
		fmt.Printf("Live reload enabled, watching: %s\n", *watch)
	}

	var uploadMaxSize int64
	var uploadDest string
	if *upload {
		var err error
		uploadMaxSize, err = parseSize(*maxSize)
		if err != nil {
			log.Fatalf("invalid max-size: %v", err)
		}
		uploadDest = *uploadDir
		if uploadDest == "" {
			uploadDest = "."
		}
		fmt.Printf("Uploads enabled (max: %s, dest: %s)\n", *maxSize, uploadDest)
	}

	if err := Serve(listenAddr, *timeout, *tlsEnabled, *certFile, *keyFile, *compress, *spa, *spaIndex, lr, *upload, uploadDest, uploadMaxSize, *zipDl); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}

func Serve(listenAddr string, timeout time.Duration, useTLS bool, cert, key string, useCompress bool, useSPA bool, spaIndexFile string, lr *LiveReload, useUpload bool, uploadDest string, uploadMaxBytes int64, useZip bool) error {
	mux := http.NewServeMux()

	if lr != nil {
		mux.Handle("/__livereload", lr)
	}

	if useUpload {
		mux.Handle("/__upload", uploadHandler(uploadDest, uploadMaxBytes))
	}

	if useZip {
		mux.Handle("/__zip", zipHandler("."))
	}

	fs := hideRootDotfiles(http.FileServer(http.Dir(".")))

	if creds != nil {
		fs = BASICAUTH(fs)
	}

	if useCompress {
		fs = gzipMiddleware(fs)
	}

	if useSPA {
		fs = spaMiddleware(fs, spaIndexFile)
	}

	if lr != nil {
		fs = liveReloadMiddleware(fs, lr)
	}

	mux.Handle("/", fs)

	svr := &http.Server{
		Addr:           listenAddr,
		Handler:        mux,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout * 2,
		IdleTimeout:    timeout * 10,
		MaxHeaderBytes: 1 << 20,
	}

	if useTLS {
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

// BASICAUTH is the basic auth middleware
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

// hideRootDotfiles middleware hides any dotfiles in the root of the directory being served
func hideRootDotfiles(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/.") {
			http.Error(w, "access to dotfiles in root directory is forbidden 😞", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AuthCreds defines the http basic authentication credentials
// Note: Though the password is not served, it is stored in plaintext
type AuthCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var creds *AuthCreds // written during  initialization

// authInit initializes basicauth
func authInit(bAuth string) error {
	if bAuth == "" {
		return nil
	}
	i := strings.Index(bAuth, ":")
	if i < 3 || i < len(bAuth)-1 {
		return errors.New("invalid basicauth flag value: value should be USERNAME:PASSWORD, e.g. dserve -basicauth admin:passw0rd")
	}
	creds = &AuthCreds{
		Username: bAuth[:i],
		Password: bAuth[i+1:],
	}
	return nil
}

// validBasicAuth checks the basicauth authentication credentials
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
