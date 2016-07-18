package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/peteretelej/dserve/dserve"
)

func main() {
	app := cli.NewApp()
	app.Name = "dserve"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "d, directory",
			Usage: "Specify the directory to be served. " +
				"Defaults to current working directory",
		},
		cli.StringFlag{
			Name: "l, listen-addr",
			Usage: "Specify listen address for the file server. " +
				"Defaults to :9011",
		},
	}
	app.Action = func(c *cli.Context) {
		dir, listenAddr := ".", ":9011"
		if len(c.Args()) > 0 {
			log.Fatal("Invalid argument passed: " + c.Args()[0])
		}
		if c.String("d") != "" {
			dir = c.String("d")
		}
		if c.String("l") != "" {
			listenAddr = c.String("l")
		}
		// Launching dserve on the current working directory
		dserve.Serve(dir, listenAddr)
	}

	app.Run(os.Args)
}
