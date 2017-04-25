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

const securedir = "secure/static"

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
