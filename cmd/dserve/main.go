package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/peteretelej/dserve"
)

var (
	dir    = flag.String("dir", "./", "the directory to serve, defaults to current directory")
	port   = flag.Int("port", 9011, "the port to serve at, defaults 9011")
	local  = flag.Bool("local", false, "whether to serve on all address or on localhost, default all addresses")
	secure = flag.Bool("secure", false, "whether to create a basic_auth secured secure/ directory, default false")
)

func main() {
	flag.Parse()

	if err := os.Chdir(*dir); err != nil {
		handleFatal(err)
		return
	}
	var addr string
	if *local {
		addr = "localhost"
	}
	listenAddr := fmt.Sprintf("%s:%d", addr, *port)

	log.Printf("Launching dserve: serving %s on %s", *dir, listenAddr)
	if err := dserve.Serve(listenAddr, *secure); err != nil {
		handleFatal(err)
		return
	}
}

func handleFatal(err error) {
	log.Print("dserve fatal error: %v", err)
	time.Sleep(5 * time.Second)
}
