package main

import (
	"log"
	"os"

	"github.com/mattjh1/psi-map/internal/cli"
)

var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	app := cli.NewApp(version, commit, buildTime)
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
