package cli

import (
	"fmt"
	"time"

	"github.com/mattjh1/psi-map/internal/logger"
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
	log := logger.GetLogger()
	start := time.Now()

	// Parse input to get URLs first (needed for URL-level cache check)
	urls, err := utils.ParseSitemap(config.Sitemap)
	if err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	log.Info("Found %d URLs to analyze", len(urls))

	// Check URL-level cache
	cachedResults, missingURLs, err := utils.CheckURLCache(config.Sitemap, urls, config.CacheTTL)
	if err != nil {
		log.Warn("Cache check failed: %v", err)
		log.Info("Continuing with full analysis")
		// Continue with full analysis if cache check fails
		missingURLs = urls
		cachedResults = nil
	}

	// Report cache status
	cachedCount := len(cachedResults)
	missingCount := len(missingURLs)

	if cachedCount > 0 {
		log.Tagged("CACHE", "Found %d cached result(s), %d URL(s) need analysis", "🎯", cachedCount, missingCount)
	} else {
		log.Tagged("CACHE", "No cached results found, analyzing all %d URLs", "📊", missingCount)
	}

	var newResults []types.PageResult

	// Only analyze missing URLs
	if missingCount > 0 {
		log.Tagged("ANALYZE", "Starting analysis of %d URL(s)...", "🔍", missingCount)
		newResults = runner.RunBatch(missingURLs, config.MaxWorkers)

		// Save new results to cache
		if err := utils.SaveURLCache(config.Sitemap, urls, newResults); err != nil {
			log.Error("Failed to save cache: %v", err)
			log.Info("Continuing...")
		} else {
			log.Tagged("CACHE", "%d new result(s) cached successfully", "💾", len(newResults))
		}
	}

	// Combine cached and new results
	allResults := combineResults(cachedResults, newResults)
	elapsed := time.Since(start)

	// Handle output based on configuration
	return handleOutput(config, allResults, elapsed)
}

// combineResults merges cached and new results, maintaining URL order from sitemap
func combineResults(cached, new []types.PageResult) []types.PageResult {
	if len(cached) == 0 {
		return new
	}
	if len(new) == 0 {
		return cached
	}

	// Create a map for quick lookup of all results by URL
	resultMap := make(map[string]types.PageResult)

	// Add cached results
	for _, result := range cached {
		resultMap[result.URL] = result
	}

	// Add new results (will overwrite any duplicates, though there shouldn't be any)
	for _, result := range new {
		resultMap[result.URL] = result
	}

	// Convert map back to slice
	combined := make([]types.PageResult, 0, len(resultMap))
	for _, result := range resultMap {
		combined = append(combined, result)
	}

	return combined
}

// handleOutput processes the results based on the configuration
func handleOutput(config *types.AnalysisConfig, results []types.PageResult, elapsed time.Duration) error {
	log := logger.GetLogger()

	switch {
	case config.StartServer:
		if err := server.Start(results, config.ServerPort); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
	case config.OutputHTML != "":
		log.Tagged("STEP", "Generating HTML report: %s", "📄", config.OutputHTML)
		if err := utils.SaveHTMLReport(results, config.OutputHTML); err != nil {
			return fmt.Errorf("failed to generate HTML report: %w", err)
		}
		log.Success("HTML report saved: %s", config.OutputHTML)
		utils.PrintSummary(results, elapsed)
	case config.OutputJSON != "":
		log.Tagged("STEP", "Generating JSON report: %s", "📋", config.OutputJSON)
		if err := utils.SaveJSONReport(results, config.OutputJSON); err != nil {
			return fmt.Errorf("failed to generate JSON report: %w", err)
		}
		log.Success("JSON report saved: %s", config.OutputJSON)
		utils.PrintSummary(results, elapsed)
	default:
		// Just print summary to console
		utils.PrintSummary(results, elapsed)
	}

	return nil
}
