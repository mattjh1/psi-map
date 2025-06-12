package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mattjh1/psi-map/internal/constants"
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
		Flags: globalFlags(),
		Action: func(c *cli.Context) error {
			return runAnalysis(c, c.Bool("server"))
		},
		Commands: []*cli.Command{
			serverCommand(),
			analyzeCommand(),
			cacheCommands(),
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
		UsageText: `psi-map [global options] command [command options] [arguments...]`,
	}
}

// globalFlags returns the global CLI flags
func globalFlags() []cli.Flag {
	defaultWorkers := max(1, runtime.NumCPU()/constants.CPUDivisor)

	return []cli.Flag{
		&cli.StringFlag{
			Name:    "html",
			Aliases: []string{"H"},
			Usage:   "Generate HTML report file (specify filename)",
		},
		&cli.StringFlag{
			Name:    "json",
			Aliases: []string{"j"},
			Usage:   "Generate JSON report file (specify filename)",
		},
		&cli.IntFlag{
			Name:    "workers",
			Aliases: []string{"w"},
			Usage:   "Maximum number of concurrent workers (default is half of available CPUs)",
			Value:   defaultWorkers,
		},
		&cli.IntFlag{
			Name:  "cache-ttl",
			Value: constants.DefaultTTLHours,
			Usage: "Cache TTL in hours (0 = no expiration)",
		},
	}
}
