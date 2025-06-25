package cli

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// NewApp creates and configures the CLI application
func NewApp(version, commit, buildTime string) *cli.App {
	return &cli.App{
		Name:        "psi-map",
		Usage:       "Analyze websites using PSI (PageSpeed Insights) API",
		Description: "A tool to analyze website overall performance using Google's PageSpeed Insights API",
		Version:     version + " (" + commit + " @ " + buildTime + ")",
		Authors: []*cli.Author{
			{
				Name:  "Mattias Holmgren",
				Email: "me@mattjh.sh",
			},
		},
		Commands: []*cli.Command{
			analyzeCommand(),
			serverCommand(),
			cacheCommands(),
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
		UsageText: `psi-map [command] [options] [arguments...]`,
	}
}
