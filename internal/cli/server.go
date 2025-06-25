package cli

import (
	"fmt"
	"runtime"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/urfave/cli/v2"
)

// serverCommand returns the server subcommand
func serverCommand() *cli.Command {
	defaultWorkers := max(1, runtime.NumCPU()/constants.CPUDivisor)

	return &cli.Command{
		Name:      "server",
		Aliases:   []string{"serve"},
		Usage:     "Start interactive web server for analysis",
		ArgsUsage: "[flags] <sitemap_url_or_file>",
		Description: `Start a web server to interactively analyze and view PageSpeed Insights results.
        
Examples:
  psi-map server sitemap.xml
  psi-map serve --port 3000 https://example.com/sitemap.xml
  psi-map serve --port 8080 sitemap.xml`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Server port",
				Value:   "8080",
			},
			&cli.IntFlag{
				Name:    "workers",
				Aliases: []string{"w"},
				Usage:   "Maximum number of concurrent workers",
				Value:   defaultWorkers,
			},
			&cli.IntFlag{
				Name:  "cache-ttl",
				Value: constants.DefaultTTLHours,
				Usage: "Cache TTL in hours (0 = no expiration)",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("sitemap URL or file path is required")
			}

			return runAnalysis(c, true)
		},
	}
}
