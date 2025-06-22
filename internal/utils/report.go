package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mattjh1/psi-map/internal/server"
	"github.com/mattjh1/psi-map/internal/types"
)

func SaveJSONReport(results []types.PageResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("failed to save JSON report %s: %w", filename, err)
	}
	return nil
}

// SaveHTMLReport generates HTML report using the server's template and functions
func SaveHTMLReport(results []types.PageResult, filename string) error {
	if err := server.GenerateHTMLFile(results, filename); err != nil {
		return fmt.Errorf("failed to generate HTML report %s: %w", filename, err)
	}
	return nil
}

// PrintSummary prints a summary to console using server's summary generation
func PrintSummary(results []types.PageResult, elapsed time.Duration) {
	summary := server.GenerateSummary(results)

	fmt.Println("\n========== SUMMARY ==========")
	fmt.Printf("Total Pages Analyzed: %d\n", summary.TotalPages)
	fmt.Printf("Successful: %d\n", summary.SuccessfulPages)
	fmt.Printf("Failed: %d\n", summary.FailedPages)

	if summary.SuccessfulPages > 0 {
		fmt.Println("\nAverage Scores:")
		if score, ok := summary.AverageScores["performance"]; ok {
			fmt.Printf("  Performance: %.1f\n", score)
		}
		if score, ok := summary.AverageScores["accessibility"]; ok {
			fmt.Printf("  Accessibility: %.1f\n", score)
		}
		if score, ok := summary.AverageScores["best_practices"]; ok {
			fmt.Printf("  Best Practices: %.1f\n", score)
		}
		if score, ok := summary.AverageScores["seo"]; ok {
			fmt.Printf("  SEO: %.1f\n", score)
		}

		fmt.Println("\nScore Distribution:")
		categories := []string{"performance", "accessibility", "best_practices", "seo"}
		for _, cat := range categories {
			if dist, ok := summary.ScoreDistribution[cat]; ok && len(dist) >= 3 {
				categoryName := formatCategoryName(cat)
				fmt.Printf("  %s: Good: %d, Needs Improvement: %d, Poor: %d\n",
					categoryName, dist[0], dist[1], dist[2])
			}
		}
	}

	fmt.Printf("\nTotal Time Elapsed: %v\n", elapsed)
	fmt.Println("=============================")
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
