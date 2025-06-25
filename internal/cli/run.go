package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/logger"
	"github.com/mattjh1/psi-map/internal/server"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils"
	"github.com/mattjh1/psi-map/runner"
	"github.com/urfave/cli/v2"
)

// runAnalysis executes the core analysis logic
func runAnalysis(c *cli.Context, isServerCommand bool) error {
	format := strings.ToLower(c.String("output"))
	outputDir := c.String("output-dir")
	name := c.String("name")

	var outputFile string
	var useStdout bool
	var outputFormat string // Add this to track the actual format

	if format == "stdout" {
		useStdout = true
		outputFormat = constants.JSON // Default format for stdout
	} else if !isServerCommand {
		// Set the output format to match the requested format
		outputFormat = format

		// Only set output file if not server command and not stdout
		var extension string
		switch format {
		case constants.JSON:
			extension = ".json"
		case constants.HTML:
			extension = ".html"
		default:
			return fmt.Errorf("unsupported output format: %s", format)
		}
		outputFile = filepath.Join(outputDir, name+extension)
	}

	config := &types.AnalysisConfig{
		Sitemap:      c.Args().First(),
		OutputFile:   outputFile,
		OutputFormat: outputFormat,
		UseStdout:    useStdout,
		StartServer:  isServerCommand,
		ServerPort:   c.String("port"),
		MaxWorkers:   c.Int("workers"),
		CacheTTL:     c.Int("cache-ttl"),
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
func combineResults(cached, fresh []types.PageResult) []types.PageResult {
	if len(cached) == 0 {
		return fresh
	}
	if len(fresh) == 0 {
		return cached
	}

	// Create a map for quick lookup of all results by URL
	resultMap := make(map[string]types.PageResult)

	// Add cached results
	for _, result := range cached {
		resultMap[result.URL] = result
	}

	// Add new results (will overwrite any duplicates, though there shouldn't be any)
	for _, result := range fresh {
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
	case config.UseStdout:
		log.Tagged("STEP", "Outputting results to stdout", "📤")
		if err := utils.SaveJSONToStdout(results); err != nil {
			return fmt.Errorf("failed to output JSON to stdout: %w", err)
		}
	case config.OutputFile != "":
		switch config.OutputFormat {
		case "html":
			log.Tagged("STEP", "Generating HTML report: %s", "📄", config.OutputFile)
			if err := utils.SaveHTMLReport(results, config.OutputFile); err != nil {
				return fmt.Errorf("failed to generate HTML report: %w", err)
			}
			log.Success("HTML report saved: %s", config.OutputFile)
		case "json":
			log.Tagged("STEP", "Generating JSON report: %s", "📋", config.OutputFile)
			if err := utils.SaveJSONReport(results, config.OutputFile); err != nil {
				return fmt.Errorf("failed to generate JSON report: %w", err)
			}
			log.Success("JSON report saved: %s", config.OutputFile)
		}
		utils.PrintSummary(results, elapsed)
	default:
		// Just print summary to console
		utils.PrintSummary(results, elapsed)
	}
	return nil
}
