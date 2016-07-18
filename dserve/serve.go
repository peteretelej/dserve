/*
package dserve launches a fileserver on that serves a specified directory on a specified listen Address. The server is a http web server open the first open port between :9012 and :9016, if the specifed address fails.
*/
package dserve

import (
	"log"
	"net/http"
	"os"
)

var basedir string

// Serve create a webserver that serves all files in a given folder(path) on a
// given listenAddress (e.g ":80") or on failure checks ":9012"-":9016"
func Serve(folder, listenAddr string) {
	if _, err := os.Stat(folder); err != nil {
		err := os.Mkdir(folder, 0700)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	basedir = folder

	// initialize secure files directory
	authInit()

	log.Println("Starting dserve on directory '" + folder + "'. Add an index.html to the folder.")
	http.Handle("/", http.FileServer(http.Dir(folder)))

	http.HandleFunc("/secure/", handleSecure)

	lAs := []string{listenAddr, ":9012", ":9013", ":9014", ":9015", ":9016"}

	for _, val := range lAs {
		log.Printf("dserve listening on %s. Visit http://localhost%s", val, val)
		err := http.ListenAndServe(val, nil)
		if err != nil {
			log.Printf("Error listening on %s. trying next port..\n", val)
		}
	}
}

func handleSecure(w http.ResponseWriter, r *http.Request) {
	if validBasicAuth(r) {
		fs := http.FileServer(http.Dir(basedir + "/secure/static"))
		h := http.StripPrefix("/secure/", fs)
		h.ServeHTTP(w, r)
		return
	}
	w.Header().Set("WWW-Authenticate", `Basic realm="Dserve secure/ Basic Authentication"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("401 Unauthorized\n"))
}
