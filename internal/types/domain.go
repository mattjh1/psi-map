package types

import (
	"time"
)

// Sitemap represents the structure of a basic XML sitemap
type Sitemap struct {
	URLs []URL `xml:"url"`
}

// URL represents a single URL entry in a sitemap
type URL struct {
	Loc string `xml:"loc"`
}

// PageResult represents the complete analysis result for a single page
// including both mobile and desktop results
type PageResult struct {
	URL      string
	Mobile   *Result
	Desktop  *Result
	Duration time.Duration
}

// GetRelevantScores returns scores from the result that determined this page's ranking
// For fastest/slowest, this would be the result with the best/worst performance
func (pr *PageResult) GetRelevantScores() *CategoryScores {
	// If both mobile and desktop exist, return scores from the one with better performance
	if pr.Mobile != nil && pr.Desktop != nil {
		if pr.Mobile.Scores != nil && pr.Desktop.Scores != nil {
			// Return scores from the result with higher performance score
			if pr.Mobile.Scores.Performance >= pr.Desktop.Scores.Performance {
				return pr.Mobile.Scores
			}
			return pr.Desktop.Scores
		}
	}

	// Fall back to whichever one exists
	if pr.Mobile != nil && pr.Mobile.Scores != nil {
		return pr.Mobile.Scores
	}
	if pr.Desktop != nil && pr.Desktop.Scores != nil {
		return pr.Desktop.Scores
	}
	return nil
}
