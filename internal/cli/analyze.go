package cli

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/urfave/cli/v2"
)

// analyzeCommand returns the analyze subcommand
func analyzeCommand() *cli.Command {
	defaultWorkers := max(1, runtime.NumCPU()/constants.CPUDivisor)

	return &cli.Command{
		Name:      "analyze",
		Aliases:   []string{"run"},
		Usage:     "Analyze sitemap and generate reports",
		ArgsUsage: "[flags] <sitemap_url_or_file>",
		Description: `Analyze a sitemap and generate reports in various formats.
        
Examples:
  psi-map analyze sitemap.xml
  psi-map analyze -o html sitemap.xml
  psi-map analyze -o json --output-dir ./reports sitemap.xml
  psi-map analyze -o stdout https://example.com/sitemap.xml`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output format: json, html, stdout (default: json)",
				Value:   constants.JSON,
			},
			&cli.StringFlag{
				Name:  "output-dir",
				Usage: "Output directory (default: current directory)",
				Value: ".",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "Output filename (without extension, default: psi-report)",
				Value: "psi-report",
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
			fmt.Printf("DEBUG: Raw CLI args: %v\n", c.Args().Slice())
			fmt.Printf("DEBUG: c.String('output'): '%s'\n", c.String("output"))
			fmt.Printf("DEBUG: c.String('o'): '%s'\n", c.String("o"))
			fmt.Printf("DEBUG: c.IsSet('output'): %t\n", c.IsSet("output"))
			fmt.Printf("DEBUG: c.IsSet('o'): %t\n", c.IsSet("o"))

			if c.NArg() < 1 {
				return fmt.Errorf("sitemap URL or file path is required")
			}

			// Handle output logic
			if err := handleOutputFlags(c); err != nil {
				return err
			}
			return runAnalysis(c, false)
		},
	}
}

func handleOutputFlags(c *cli.Context) error {
	format := strings.ToLower(c.String("output"))

	// Validate format
	switch format {
	case constants.STDOUT, constants.HTML, constants.JSON:
		// Valid formats
	default:
		return fmt.Errorf("unsupported output format: %s (supported: json, html, stdout)", format)
	}

	return nil
}
