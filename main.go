package main

import (
	"os"

	"github.com/mattjh1/psi-map/internal/cli"
	"github.com/mattjh1/psi-map/internal/logger"
)

var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	logger.Init(logger.WithOutput(os.Stderr))
	app := cli.NewApp(version, commit, buildTime)
	if err := app.Run(os.Args); err != nil {
		logger.GetLogger().Error("%v", err)
		os.Exit(1)
	}
}
