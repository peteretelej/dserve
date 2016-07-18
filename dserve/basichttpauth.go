package dserve

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// AuthCreds defines the http basic authentication credentials for /secure
// Note: Though the password is not served, it is stored in plaintext
type AuthCreds struct {
	invalid  string
	Username string `json:"username"`
	Password string `json:"password"`
}

var securedir string

// authInit initializes the secure directory
func authInit() {
	// files served on /secure (secure/static)
	securedir = basedir + "/secure/static"
	if _, err := os.Stat(securedir); err != nil {
		err := os.MkdirAll(securedir, 0700)
		if err != nil {
			log.Fatal(err.Error())
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
		err = ioutil.WriteFile(basedir+"/secure/securepass.json.sample", d, 0644)
	}
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
	sp, err := ioutil.ReadFile(basedir + "/secure/securepass.json")
	if err != nil {
		return creds, err
	}
	err = json.Unmarshal(sp, &creds)
	if err != nil {
		return creds, err
	}
	if creds.Username == "" && creds.Password == "" {
		return creds, errors.New("No username and password in securepass.json.")
	}
	return creds, nil
}
