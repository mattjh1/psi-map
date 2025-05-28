// main.go
package main

import (
	"log"
	"os"

	"github.com/mattjh1/psi-map/internal/cli"
)

func main() {
	app := cli.NewApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
