package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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
	secure    = flag.Bool("secure", false, "whether to create a basic_auth secured secure/ directory, default false")
	basicauth = flag.String("basicauth", ".basicauth.json", "file to be used for basicauth json config")
	timeout   = flag.Duration("timeout", time.Minute*3, "http server read timeout, write timeout will be double this")
)

var usage = func() {
	fmt.Fprintf(os.Stderr, `dserve serves a static directory over http

Usage:
	dserve
	dserve [flags].. [directory]

Examples:
	dserve 			Serves the current directory over http at :9011
	dserve -local		Serves the current directory on localhost:9011
	dserve -dir ~/dir	Serves the directory ~/dir over http 
	dserve -secure		Serves the current directory with basicauth using sample .basicauth.json
	dserve -secure -basicauth myauth.json
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
	if *secure {
		if err := authInit(); err != nil {
			fmt.Printf("Basic Auth credentials %s missing: edit and rename %s.sample\n",
				*basicauth, *basicauth)
			os.Exit(1)
		}
	}

	listenAddr := fmt.Sprintf("%s:%d", addr, *port)
	fmt.Printf("Launching dserve http server %s on %s\n", *dir, listenAddr)
	if err := Serve(listenAddr, *secure, *timeout); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}

// Serve launches HTTP server serving on listenAddr and servers a basic_auth secured directory at secure/static
func Serve(listenAddr string, secureDir bool, timeout time.Duration) error {
	mux := http.NewServeMux()

	fs := hideRootDotfiles(http.FileServer(http.Dir(".")))

	switch secureDir {
	case true:
		if err := authInit(); err != nil {
			return fmt.Errorf("failed to initialize basic auth: %v", err)
		}
		fmt.Printf("BasicAuth enabled using credentials in %s\n", *basicauth)
		mux.Handle("/", BASICAUTH(fs))
	default:
		mux.Handle("/", fs)
	}

	svr := &http.Server{
		Addr:           listenAddr,
		Handler:        mux,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout * 2,
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
	invalid  string
	Username string `json:"username"`
	Password string `json:"password"`
}

// authInit initializes the secure directory
func authInit() error {
	// get creds
	err := func() error {
		// Read the securepass.json creds
		_, err := getCreds()
		return err
	}()
	if err != nil {
		sample := &AuthCreds{Username: "example", Password: "pass123"}
		d, err := json.MarshalIndent(sample, "", "	")
		if err != nil {
			log.Print(err)
			return fmt.Errorf("internal error")
		}
		if err := ioutil.WriteFile(fmt.Sprintf("%s", *basicauth), d, 0644); err != nil {
			return fmt.Errorf("unable to create sample file %s", *basicauth)
		}
	}
	return nil
}

// validBasicAuth checks the basicauth authentication credentials
func validBasicAuth(r *http.Request) bool {
	creds, err := getCreds()
	if err != nil {
		return false
	}
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}
	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}
	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}
	return pair[0] == creds.Username && pair[1] == creds.Password
}

// getCreds gets the current http basic credentials
func getCreds() (*AuthCreds, error) {
	creds := &AuthCreds{}
	sp, err := ioutil.ReadFile(*basicauth)
	if err != nil {
		return creds, err
	}
	err = json.Unmarshal(sp, &creds)
	if err != nil {
		return creds, err
	}
	if creds.Username == "" && creds.Password == "" {
		return creds, fmt.Errorf("no username and password in %s", *basicauth)
	}
	return creds, nil
}
