package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mattjh1/psi-map/internal/types"
)

func TestGenerateSummary_EmptyResults(t *testing.T) {
	server := &Server{results: []types.PageResult{}}

	summary := server.generateSummary()

	assert.Equal(t, 0, summary.TotalPages)
	assert.Equal(t, 0, summary.SuccessfulPages)
	assert.Equal(t, 0, summary.FailedPages)
	assert.Empty(t, summary.AverageScores)
	assert.Len(t, summary.ScoreDistribution, 4) // should have 4 categories initialized

	// Each category should have 3 buckets initialized to 0
	for _, category := range []string{"performance", "accessibility", "best_practices", "seo"} {
		assert.Len(t, summary.ScoreDistribution[category], 3)
		assert.Equal(t, []int{0, 0, 0}, summary.ScoreDistribution[category])
	}
}

func TestGenerateSummary_AllSuccessfulResults(t *testing.T) {
	results := []types.PageResult{
		createMockResult("https://example1.com", 95, 90, 85, 98, false), // good scores
		createMockResult("https://example2.com", 75, 80, 70, 85, false), // mixed scores
		createMockResult("https://example3.com", 45, 50, 40, 55, false), // poor scores
	}

	server := &Server{results: results}
	summary := server.generateSummary()

	assert.Equal(t, 3, summary.TotalPages)
	assert.Equal(t, 3, summary.SuccessfulPages)
	assert.Equal(t, 0, summary.FailedPages)

	// Check average calculations (each result contributes mobile + desktop scores)
	expectedPerformance := (95 + 95 + 75 + 75 + 45 + 45) / 6.0 // 71.67
	assert.InDelta(t, expectedPerformance, summary.AverageScores["performance"], 0.01)

	expectedAccessibility := (90 + 90 + 80 + 80 + 50 + 50) / 6.0 // 73.33
	assert.InDelta(t, expectedAccessibility, summary.AverageScores["accessibility"], 0.01)

	// Check score distributions
	// 95,95 (good), 75,75 (needs improvement), 45,45 (poor)
	assert.Equal(t, []int{2, 2, 2}, summary.ScoreDistribution["performance"])
	// 90,90 (good), 80,80,50,50 (needs improvement)
	assert.Equal(t, []int{2, 4, 0}, summary.ScoreDistribution["accessibility"])
}

func TestGenerateSummary_MixedSuccessAndErrors(t *testing.T) {
	results := []types.PageResult{
		createMockResult("https://example1.com", 90, 85, 80, 95, false), // success
		createMockResult("https://example2.com", 70, 75, 65, 80, true),  // both mobile and desktop error
		createMockResult("https://example3.com", 40, 45, 50, 55, false), // success
	}

	server := &Server{results: results}
	summary := server.generateSummary()

	assert.Equal(t, 3, summary.TotalPages)
	assert.Equal(t, 2, summary.SuccessfulPages)
	assert.Equal(t, 1, summary.FailedPages)

	// Only successful results should contribute to averages
	expectedPerformance := (90 + 90 + 40 + 40) / 4.0 // 65.0
	assert.InDelta(t, expectedPerformance, summary.AverageScores["performance"], 0.01)
}

func TestGenerateSummary_PartialSuccessResults(t *testing.T) {
	results := []types.PageResult{
		createResultWithPartialSuccess("https://example1.com", true, false),  // mobile success, desktop error
		createResultWithPartialSuccess("https://example2.com", false, true),  // mobile error, desktop success
		createResultWithPartialSuccess("https://example3.com", false, false), // both error
	}

	server := &Server{results: results}
	summary := server.generateSummary()

	assert.Equal(t, 3, summary.TotalPages)
	assert.Equal(t, 2, summary.SuccessfulPages) // first two should be successful (at least one device worked)
	assert.Equal(t, 1, summary.FailedPages)     // only third should be failed

	// Should include scores from successful mobile (80) and successful desktop (85)
	expectedPerformance := (80 + 85) / 2.0 // 82.5
	assert.InDelta(t, expectedPerformance, summary.AverageScores["performance"], 0.01)
}

func TestGenerateSummary_ZeroScores(t *testing.T) {
	results := []types.PageResult{
		createMockResult("https://example1.com", 0, 85, 90, 95, false), // zero performance score
		createMockResult("https://example2.com", 90, 0, 80, 85, false), // zero accessibility score
	}

	server := &Server{results: results}
	summary := server.generateSummary()

	// Zero scores should be excluded from averages
	expectedPerformance := (90 + 90) / 2.0 // Only second result's mobile+desktop
	assert.InDelta(t, expectedPerformance, summary.AverageScores["performance"], 0.01)

	expectedAccessibility := (85 + 85) / 2.0 // Only first result's mobile+desktop
	assert.InDelta(t, expectedAccessibility, summary.AverageScores["accessibility"], 0.01)

	// Zero scores should not contribute to distributions
	assert.Equal(t, []int{2, 0, 0}, summary.ScoreDistribution["performance"])   // 2 good scores (90,90)
	assert.Equal(t, []int{0, 2, 0}, summary.ScoreDistribution["accessibility"]) // 2 need improvement scores (85,85)
}

func TestProcessScores(t *testing.T) {
	server := &Server{}
	totalScores := make(map[string]float64)
	scoreCounts := make(map[string]int)
	distribution := map[string][]int{
		"performance":    {0, 0, 0},
		"accessibility":  {0, 0, 0},
		"best_practices": {0, 0, 0},
		"seo":            {0, 0, 0},
	}

	scores := &types.CategoryScores{
		Performance:   95, // good
		Accessibility: 75, // needs improvement
		BestPractices: 45, // poor
		SEO:           0,  // should be ignored
	}

	server.processScores(scores, totalScores, scoreCounts, distribution)

	// Check totals and counts
	assert.Equal(t, 95.0, totalScores["performance"])
	assert.Equal(t, 75.0, totalScores["accessibility"])
	assert.Equal(t, 45.0, totalScores["best_practices"])
	assert.Equal(t, 0.0, totalScores["seo"]) // zero score ignored

	assert.Equal(t, 1, scoreCounts["performance"])
	assert.Equal(t, 1, scoreCounts["accessibility"])
	assert.Equal(t, 1, scoreCounts["best_practices"])
	assert.Equal(t, 0, scoreCounts["seo"]) // zero score ignored

	// Check distributions
	assert.Equal(t, []int{1, 0, 0}, distribution["performance"])    // good bucket
	assert.Equal(t, []int{0, 1, 0}, distribution["accessibility"])  // needs improvement bucket
	assert.Equal(t, []int{0, 0, 1}, distribution["best_practices"]) // poor bucket
	assert.Equal(t, []int{0, 0, 0}, distribution["seo"])            // zero score ignored
}

func TestGenerateSummary_FastestAndSlowestPages(t *testing.T) {
	results := []types.PageResult{
		{
			URL:      "https://fast.com",
			Duration: 50 * time.Millisecond,
			Mobile:   types.Result{Scores: &types.CategoryScores{Performance: 90}, Error: nil},
			Desktop:  types.Result{Scores: &types.CategoryScores{Performance: 95}, Error: nil},
		},
		{
			URL:      "https://slow.com",
			Duration: 500 * time.Millisecond,
			Mobile:   types.Result{Scores: &types.CategoryScores{Performance: 80}, Error: nil},
			Desktop:  types.Result{Scores: &types.CategoryScores{Performance: 85}, Error: nil},
		},
		{
			URL:      "https://medium.com",
			Duration: 200 * time.Millisecond,
			Mobile:   types.Result{Scores: &types.CategoryScores{Performance: 75}, Error: nil},
			Desktop:  types.Result{Scores: &types.CategoryScores{Performance: 80}, Error: nil},
		},
	}

	server := &Server{results: results}
	summary := server.generateSummary()

	// Note: The current implementation assigns the Mobile result as reference
	// This is a bit odd but testing current behavior
	assert.NotNil(t, summary.FastestPage)
	assert.NotNil(t, summary.SlowestPage)

	// The fastest page should have performance score 90 (mobile score from fast.com)
	assert.Equal(t, 90.0, summary.FastestPage.Scores.Performance)

	// The slowest page should have performance score 80 (mobile score from slow.com)
	assert.Equal(t, 80.0, summary.SlowestPage.Scores.Performance)
}

func TestGenerateSummary_PublicFunction(t *testing.T) {
	results := []types.PageResult{
		createMockResult("https://example1.com", 90, 85, 80, 95, false),
		createMockResult("https://example2.com", 70, 75, 65, 80, false),
	}

	summary := GenerateSummary(results)

	assert.Equal(t, 2, summary.TotalPages)
	assert.Equal(t, 2, summary.SuccessfulPages)
	assert.Equal(t, 0, summary.FailedPages)

	expectedPerformance := (90 + 90 + 70 + 70) / 4.0 // 80.0
	assert.InDelta(t, expectedPerformance, summary.AverageScores["performance"], 0.01)
}
