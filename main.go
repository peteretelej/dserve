package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
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
	dir     = flag.String("dir", "./", "the directory to serve, defaults to current directory")
	port    = flag.Int("port", 9011, "the port to serve at, defaults 9011")
	local   = flag.Bool("local", false, "whether to serve on all address or on localhost, default all addresses")
	secure  = flag.Bool("secure", false, "whether to create a basic_auth secured secure/ directory, default false")
	timeout = flag.Duration("timeout", time.Minute*3, "http server read timeout, write timeout will be double this")
)

const securedir = "secure/static"

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

	listenAddr := fmt.Sprintf("%s:%d", addr, *port)
	fmt.Printf("Launching dserve: serving %s on %s\n", *dir, listenAddr)
	if err := Serve(listenAddr, *secure, *timeout); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}

// Serve launches HTTP server serving on listenAddr and servers a basic_auth secured directory at secure/static
func Serve(listenAddr string, secureDir bool, timeout time.Duration) error {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("."))

	switch secureDir {
	case true:
		if err := authInit(); err != nil {
			return fmt.Errorf("failed to initialize basic auth: %v", err)
		}
		mux.Handle("/", BASICAUTH(http.StripPrefix("/secure/", fs)))
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

// BASICAUTH is the basic auth handler
func BASICAUTH(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validBasicAuth(r) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Dserve secure/ Basic Authentication"`)
			http.Error(w, "Not Authorized", http.StatusUnauthorized)
			return
		}
		return next.ServeHTTP(w, r)
	})
}

// AuthCreds defines the http basic authentication credentials for /secure
// Note: Though the password is not served, it is stored in plaintext
type AuthCreds struct {
	invalid  string
	Username string `json:"username"`
	Password string `json:"password"`
}

// authInit initializes the secure directory
func authInit() error {
	if _, err := os.Stat(securedir); err != nil {
		err := os.MkdirAll(securedir, 0700)
		if err != nil {
			return err
		}
	}
	// get creds
	err := func() error {
		// Read the securepass.json creds
		_, err := getCreds()
		return err
	}()
	if err != nil {
		// create sample securepass.json example
		sample := &AuthCreds{Username: "example", Password: "pass123"}
		d, err := json.MarshalIndent(sample, "", "	")
		if err != nil {
			log.Fatal(err.Error())
		}
		return ioutil.WriteFile("secure/securepass.json.sample", d, 0644)
	}
	return nil
}

// validBasicAuth checks the authentication credentials to access /secure files
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
	sp, err := ioutil.ReadFile("/secure/securepass.json")
	if err != nil {
		return creds, err
	}
	err = json.Unmarshal(sp, &creds)
	if err != nil {
		return creds, err
	}
	if creds.Username == "" && creds.Password == "" {
		return creds, errors.New("no username and password in securepass.json")
	}
	return creds, nil
}
