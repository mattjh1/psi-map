package cli

import (
	"fmt"
	"time"

	"github.com/mattjh1/psi-map/internal/server"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils"
	"github.com/mattjh1/psi-map/runner"
	"github.com/urfave/cli/v2"
)

// runAnalysis executes the core analysis logic
func runAnalysis(c *cli.Context, forceServer bool) error {
	config := &types.AnalysisConfig{
		Sitemap:     c.String("sitemap"),
		OutputHTML:  c.String("html"),
		OutputJSON:  c.String("json"),
		StartServer: c.Bool("server") || forceServer,
		ServerPort:  c.String("port"),
		MaxWorkers:  c.Int("workers"),
		CacheTTL:    c.Int("cache-ttl"),
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
		fmt.Println("[INFO] Results cached successfully")
	}

	// Handle output based on configuration
	return handleOutput(config, results, elapsed)
}

// handleOutput processes the results based on the configuration
func handleOutput(config *types.AnalysisConfig, results []types.PageResult, elapsed time.Duration) error {
	switch {
	case config.StartServer:
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
