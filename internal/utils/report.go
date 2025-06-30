package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mattjh1/psi-map/internal/logger"
	"github.com/mattjh1/psi-map/internal/server"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils/validate"
)

// serverGenerateHTMLFile is a package-level variable to allow mocking in tests
var serverGenerateHTMLFile = server.GenerateHTMLFile

func SaveJSONReport(results []*types.PageResult, filename string) error {
	components := validate.SplitFilePath(filename)

	// Use the secure file creation function
	file, _, err := validate.SafeCreateFile(components.Dir, components.Name, components.Extension)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("failed to save JSON report %s: %w", filename, err)
	}
	return nil
}

// SaveJSONToStdout outputs JSON results to stdout for piping
func SaveJSONToStdout(results []*types.PageResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("failed to encode JSON to stdout: %w", err)
	}
	return nil
}

// SaveHTMLReport generates HTML report using the server's template and functions
func SaveHTMLReport(results []*types.PageResult, filename string) error {
	if err := serverGenerateHTMLFile(results, filename); err != nil {
		return fmt.Errorf("failed to generate HTML report %s: %w", filename, err)
	}
	return nil
}

// PrintSummary prints a summary to console using server's summary generation
func PrintSummary(results []*types.PageResult, elapsed time.Duration) {
	log := logger.GetLogger()
	ui := log.UI()
	summary := server.GenerateSummary(results)

	ui.Header("SUMMARY")
	log.Info("Total Pages Analyzed: %d", summary.TotalPages)
	log.Success("Successful: %d", summary.SuccessfulPages)
	log.Error("Failed: %d", summary.FailedPages)

	if summary.SuccessfulPages > 0 {
		ui.Section("Average Scores")
		if score, ok := summary.AverageScores["performance"]; ok {
			log.Info("  Performance: %.1f", score)
		}
		if score, ok := summary.AverageScores["accessibility"]; ok {
			log.Info("  Accessibility: %.1f", score)
		}
		if score, ok := summary.AverageScores["best_practices"]; ok {
			log.Info("  Best Practices: %.1f", score)
		}
		if score, ok := summary.AverageScores["seo"]; ok {
			log.Info("  SEO: %.1f", score)
		}

		ui.Section("Score Distribution")
		categories := []string{"performance", "accessibility", "best_practices", "seo"}
		for _, cat := range categories {
			if dist, ok := summary.ScoreDistribution[cat]; ok && len(dist) >= 3 {
				categoryName := formatCategoryName(cat)
				log.Info("  %s: Good: %d, Needs Improvement: %d, Poor: %d",
					categoryName, dist[0], dist[1], dist[2])
			}
		}
	}

	log.Info("Total Time Elapsed: %v", elapsed)
}

// formatCategoryName converts snake_case to Title Case
func formatCategoryName(s string) string {
	switch s {
	case "performance":
		return "Performance"
	case "accessibility":
		return "Accessibility"
	case "best_practices":
		return "Best Practices"
	case "seo":
		return "SEO"
	default:
		return s
	}
}
