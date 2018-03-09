package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	dir       = flag.String("dir", "./", "the directory to serve, defaults to current directory")
	port      = flag.Int("port", 9011, "the port to serve at, defaults 9011")
	local     = flag.Bool("local", false, "whether to serve on all address or on localhost, default all addresses")
	basicauth = flag.String("basicauth", "", "enable basic authentication")
	timeout   = flag.Duration("timeout", time.Minute*3, "http server read timeout, write timeout will be double this")
)

var usage = func() {
	fmt.Fprint(os.Stderr, `dserve serves a static directory over http

Usage:
	dserve
	dserve [flags]..

Examples:
	dserve 			Serves the current directory over http at :9011
	dserve -local		Serves the current directory on localhost:9011
	dserve -dir ~/dir	Serves the directory ~/dir over http 
	dserve -basicauth admin:Passw0rd
				Serves the current directory with basicauth using config file myauth.json

Flags:
`)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
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
		fmt.Print("invalid basicauth flag value: value should be USERNAME:PASSWORD, e.g. dserve -basicauth admin:passw0rd")
		os.Exit(1)
	}

	listenAddr := fmt.Sprintf("%s:%d", addr, *port)
	fmt.Printf("Launching dserve http server %s on %s\n", *dir, listenAddr)
	if err := Serve(listenAddr, *timeout); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}

// Serve launches HTTP server serving on listenAddr and servers a basic_auth secured directory at secure/static
func Serve(listenAddr string, timeout time.Duration) error {
	mux := http.NewServeMux()

	fs := hideRootDotfiles(http.FileServer(http.Dir(".")))

	if creds != nil {
		fs = BASICAUTH(fs)
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
	if i < 3 && i < len(bAuth)-1 {
		return fmt.Errorf("invalid basicauth flag value")
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
