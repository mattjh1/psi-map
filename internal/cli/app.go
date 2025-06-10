package cli

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mattjh1/psi-map/internal/server"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils"
	"github.com/mattjh1/psi-map/runner"
	"github.com/urfave/cli/v2"
)

// NewApp creates and configures the CLI application
func NewApp() *cli.App {
	return &cli.App{
		Name:        "psi-map",
		Usage:       "Analyze websites using PSI (PageSpeed Insights) API",
		Description: "A tool to analyze website overall performance using Google's PageSpeed Insights API",
		Version:     "0.1.0",
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
			cacheCommand(),
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
		UsageText: `psi-map [global options] command [command options] [arguments...]

EXAMPLES:
   psi-map --sitemap https://example.com/sitemap.xml --html report.html
   psi-map --sitemap sitemap.xml --server --port 8080
   psi-map --sitemap https://example.com/sitemap.xml --json results.json
   psi-map analyze --sitemap https://example.com/sitemap.xml --html report.html --workers 10
   psi-map server --sitemap sitemap.xml --port 9000
   psi-map cache list
   psi-map cache clear`,
	}
}

// globalFlags returns the global CLI flags
func globalFlags() []cli.Flag {
	defaultWorkers := max(1, runtime.NumCPU()/2)

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
			Value: 24,
			Usage: "Cache TTL in hours (0 = no expiration)",
		},
	}
}

// serverCommand returns the server subcommand
func serverCommand() *cli.Command {
	return &cli.Command{
		Name:    "server",
		Aliases: []string{"serve"},
		Usage:   "Start web server with analysis results",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "sitemap",
				Aliases:  []string{"s"},
				Usage:    "URL or sitemap.xml file path",
				Required: true,
			},
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
				Value:   5,
			},
		},
		Action: func(c *cli.Context) error {
			return runAnalysis(c, true)
		},
	}
}

// analyzeCommand returns the analyze subcommand
func analyzeCommand() *cli.Command {
	return &cli.Command{
		Name:    "analyze",
		Aliases: []string{"run"},
		Usage:   "Analyze sitemap and generate reports",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "sitemap",
				Aliases:  []string{"s"},
				Usage:    "URL or sitemap.xml file path",
				Required: true,
			},
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
				Usage:   "Maximum number of concurrent workers",
				Value:   5,
			},
		},
		Action: func(c *cli.Context) error {
			return runAnalysis(c, false)
		},
	}
}

// Add cache management commands
func cacheCommand() *cli.Command {
	return &cli.Command{
		Name:  "cache",
		Usage: "Cache management commands",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List cached results",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "ttl",
						Value: 24,
						Usage: "TTL in hours for expiration check",
					},
				},
				Action: func(c *cli.Context) error {
					ttl := c.Int("ttl")
					cacheInfos, err := utils.ListCacheFiles(ttl)
					if err != nil {
						return err
					}

					if len(cacheInfos) == 0 {
						fmt.Println("No cached results found")
						return nil
					}

					fmt.Printf("Found %d cached result(s) (TTL: %dh):\n\n", len(cacheInfos), ttl)
					fmt.Printf("%-12s %-8s %-50s %s\n", "AGE", "STATUS", "SITEMAP", "HASH")
					fmt.Println(strings.Repeat("-", 90))

					for _, info := range cacheInfos {
						status := "VALID"
						if info.IsExpired {
							status = "EXPIRED"
						}

						sitemap := info.SitemapURL
						if len(sitemap) > 45 {
							sitemap = "..." + sitemap[len(sitemap)-42:]
						}

						fmt.Printf("%-12s %-8s %-50s %s\n",
							info.Age, status, sitemap, info.Hash[:8]+"...")
					}
					return nil
				},
			},
			{
				Name:  "clean",
				Usage: "Remove expired cache files",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "ttl",
						Value: 24,
						Usage: "TTL in hours",
					},
				},
				Action: func(c *cli.Context) error {
					ttl := c.Int("ttl")
					cleaned, err := utils.CleanExpiredCache(ttl)
					if err != nil {
						return err
					}

					if cleaned == 0 {
						fmt.Printf("No expired cache files found (TTL: %dh)\n", ttl)
					} else {
						fmt.Printf("Cleaned %d expired cache file(s) (TTL: %dh)\n", cleaned, ttl)
					}
					return nil
				},
			},
			{
				Name:  "clear",
				Usage: "Clear all cached results",
				Action: func(c *cli.Context) error {
					if err := utils.ClearCache(); err != nil {
						return err
					}
					fmt.Println("Cache cleared successfully")
					return nil
				},
			},
		},
	}
}

// runAnalysis executes the core analysis logic
func runAnalysis(c *cli.Context, forceServer bool) error {
	config := &types.AnalysisConfig{
		Sitemap:     c.String("sitemap"),
		OutputHTML:  c.String("html"),
		OutputJSON:  c.String("json"),
		StartServer: c.Bool("server") || forceServer,
		ServerPort:  c.String("port"),
		MaxWorkers:  c.Int("workers"),
	}

	return executeAnalysis(config)
}

// executeAnalysis runs the analysis with the given configuration
func executeAnalysis(config *types.AnalysisConfig) error {
	start := time.Now()

	results, found, err := utils.CheckCache(config.Sitemap, config.CacheTTL)
	if err != nil {
		fmt.Printf("cache check failed: %v", err)
	}

	if found {
		fmt.Println("Using cached results for sitemap")
		elapsed := time.Since(start)
		return handleOutput(config, results, elapsed)
	}

	fmt.Printf("No cache found, starting analysis of: %s\n", config.Sitemap)

	// Parse input to get URLs
	urls, err := utils.ParseSitemap(config.Sitemap)
	if err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	fmt.Printf("Found %d URLs to analyze\n", len(urls))

	// Run analysis using the runner
	results = runner.RunBatch(urls, config.MaxWorkers)
	elapsed := time.Since(start)

	if err := utils.SaveCache(config.Sitemap, results); err != nil {
		fmt.Printf("Failed to save cache: %v", err)
		fmt.Printf("Continuing...")
	} else {
		fmt.Printf("Results cached successfully")
	}

	// Handle output based on configuration
	return handleOutput(config, results, elapsed)
}

// handleOutput processes the results based on the configuration
func handleOutput(config *types.AnalysisConfig, results []types.PageResult, elapsed time.Duration) error {
	switch {
	case config.StartServer:
		fmt.Println("Starting web server...")
		if err := server.Start(results, config.ServerPort); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
	case config.OutputHTML != "":
		fmt.Printf("Generating HTML report: %s\n", config.OutputHTML)
		if err := utils.SaveHTMLReport(results, config.OutputHTML); err != nil {
			return fmt.Errorf("failed to generate HTML report: %w", err)
		}
		fmt.Printf("HTML report saved: %s\n", config.OutputHTML)
		utils.PrintSummary(results, elapsed)
	case config.OutputJSON != "":
		fmt.Printf("Generating JSON report: %s\n", config.OutputJSON)
		if err := utils.SaveJSONReport(results, config.OutputJSON); err != nil {
			return fmt.Errorf("failed to generate JSON report: %w", err)
		}
		fmt.Printf("JSON report saved: %s\n", config.OutputJSON)
		utils.PrintSummary(results, elapsed)
	default:
		// Just print summary to console
		utils.PrintSummary(results, elapsed)
	}

	return nil
}
