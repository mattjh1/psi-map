package server

import (
	"time"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/types"
)

// generateSummary creates aggregate statistics from results
func (s *Server) generateSummary() types.ReportSummary {
	summary := types.ReportSummary{
		TotalPages:        len(s.results),
		AverageScores:     make(map[string]float64),
		ScoreDistribution: make(map[string][]int),
	}

	var (
		totalScores = make(map[string]float64)
		scoreCounts = make(map[string]int)
		fastestTime = time.Hour
		slowestTime time.Duration
	)

	categories := []string{"performance", "accessibility", "best_practices", "seo"}
	for _, cat := range categories {
		summary.ScoreDistribution[cat] = []int{0, 0, 0}
	}

	for _, pageResult := range s.results {
		if pageResult == nil {
			continue // Skip nil page results
		}

		// Check if either mobile or desktop succeeded
		mobileSuccess := pageResult.Mobile != nil && pageResult.Mobile.Error == nil
		desktopSuccess := pageResult.Desktop != nil && pageResult.Desktop.Error == nil

		if mobileSuccess || desktopSuccess {
			summary.SuccessfulPages++
		} else {
			summary.FailedPages++
			continue
		}

		// Track timing
		if pageResult.Duration < fastestTime {
			fastestTime = pageResult.Duration
			summary.FastestPage = pageResult
		}
		if pageResult.Duration > slowestTime {
			slowestTime = pageResult.Duration
			summary.SlowestPage = pageResult
		}

		// Process mobile scores if available
		if pageResult.Mobile != nil && pageResult.Mobile.Error == nil && pageResult.Mobile.Scores != nil {
			s.processScores(pageResult.Mobile.Scores, totalScores, scoreCounts, summary.ScoreDistribution)
		}

		// Process desktop scores if available
		if pageResult.Desktop != nil && pageResult.Desktop.Error == nil && pageResult.Desktop.Scores != nil {
			s.processScores(pageResult.Desktop.Scores, totalScores, scoreCounts, summary.ScoreDistribution)
		}
	}

	// Calculate averages
	for category, total := range totalScores {
		if count := scoreCounts[category]; count > 0 {
			summary.AverageScores[category] = total / float64(count)
		}
	}

	return summary
}

// processScores is a helper function to process scores
func (s *Server) processScores(scores *types.CategoryScores, totalScores map[string]float64, scoreCounts map[string]int, distribution map[string][]int) {
	scoreMap := map[string]float64{
		"performance":    scores.Performance,
		"accessibility":  scores.Accessibility,
		"best_practices": scores.BestPractices,
		"seo":            scores.SEO,
	}

	for category, score := range scoreMap {
		if score <= 0 {
			continue
		}
		totalScores[category] += score
		scoreCounts[category]++

		var bucket int
		switch {
		case score >= constants.ScoreGoodThreshold:
			bucket = 0 // good
		case score >= constants.ScorePoorThreshold:
			bucket = 1 // needs improvement
		default:
			bucket = 2 // poor
		}
		distribution[category][bucket]++
	}
}

// GenerateSummary creates a summary from results without needing a server instance
func GenerateSummary(results []*types.PageResult) types.ReportSummary {
	s := &Server{results: results}
	return s.generateSummary()
}
