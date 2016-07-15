/*
package dserve launches a fileserver on that serves a specified directory on a specified listen Address. The server is a http web server open the first open port between :9012 and :9016, if the specifed address fails.
*/
package dserve

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

// Serve create a webserver that serves all files in a given folder(path) on a
// given listenAddress (e.g ":80") or on failure checks ":9012"-":9016"
func Serve(folder, listenAddr string) error {
	if _, err := os.Stat(folder); err != nil {
		err := os.Mkdir(folder, 0700)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(path.Join(folder, "index.html")); err != nil {
		indexhtml := []byte(indexHtml)
		err = ioutil.WriteFile(path.Join(folder, "index.html"), indexhtml, 0644)
		if err != nil {
			return errors.New("Could not find index.html in folder, " +
				"and unable to write file: " + err.Error())
		}
	}
	log.Println("Starting dserve on directory '" + folder + "'")
	http.Handle("/", http.FileServer(http.Dir(folder)))

	lAs := []string{listenAddr, ":9012", ":9013", ":9014", ":9015", ":9016"}

	for _, val := range lAs {
		log.Printf("dserve listening on %s. Visit http://localhost%s", val, val)
		err := http.ListenAndServe(val, nil)
		if err != nil {
			log.Printf("Error listening on %s. trying next port..\n", val)
		}
	}

	return nil
}
