package cli

import (
	"testing"
	"time"

	"github.com/mattjh1/psi-map/internal/cache"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a CategoryScores object
func makeScores(p, a, bp, seo float64) *types.CategoryScores {
	return &types.CategoryScores{
		Performance:   p,
		Accessibility: a,
		BestPractices: bp,
		SEO:           seo,
	}
}

// Helper to create a PageResult with Mobile and Desktop results
func createMockResult(url string, mobileScores, desktopScores *types.CategoryScores) *types.PageResult {
	if mobileScores == nil {
		mobileScores = &types.CategoryScores{} // zeroed scores, non-nil
	}
	if desktopScores == nil {
		desktopScores = &types.CategoryScores{}
	}
	return &types.PageResult{
		URL: url,
		Mobile: &types.Result{
			URL:     url,
			Scores:  mobileScores,
			Elapsed: 100 * time.Millisecond,
		},
		Desktop: &types.Result{
			URL:     url,
			Scores:  desktopScores,
			Elapsed: 150 * time.Millisecond,
		},
		Duration: 250 * time.Millisecond,
	}
}

func TestCombineResultsInOrder(t *testing.T) {
	orderedURLs := []string{"https://cached.com", "https://fresh.com"}
	cached := []*types.PageResult{
		createMockResult("https://cached.com", makeScores(0.8, 0.9, 1, 0.95), makeScores(0.4, 0.7, 1, 1)),
	}
	fresh := []*types.PageResult{
		createMockResult("https://fresh.com", makeScores(0.9, 0.9, 0.9, 0.9), makeScores(0.4, 0.7, 0.5, 0.8)),
	}
	result := cache.CombineResultsInOrder(orderedURLs, cached, fresh)
	require.Len(t, result, 2)
	assert.Equal(t, "https://cached.com", result[0].URL)
	assert.Equal(t, "https://fresh.com", result[1].URL)
	assert.InDelta(t, 0.8, result[0].Mobile.Scores.Performance, 0.01)
	assert.InDelta(t, 0.9, result[1].Mobile.Scores.Performance, 0.01)
}

func TestCombineResultsInOrder_OnlyCached(t *testing.T) {
	orderedURLs := []string{"https://onlycached.com"}
	cached := []*types.PageResult{
		createMockResult("https://onlycached.com", makeScores(0.7, 0.8, 0.9, 1), nil),
	}
	result := cache.CombineResultsInOrder(orderedURLs, cached, nil)
	require.Len(t, result, 1)
	assert.Equal(t, "https://onlycached.com", result[0].URL)
}

func TestCombineResultsInOrder_OnlyFresh(t *testing.T) {
	orderedURLs := []string{"https://onlyfresh.com"}
	fresh := []*types.PageResult{
		createMockResult("https://onlyfresh.com", makeScores(1, 1, 1, 1), nil),
	}
	result := cache.CombineResultsInOrder(orderedURLs, nil, fresh)
	require.Len(t, result, 1)
	assert.Equal(t, "https://onlyfresh.com", result[0].URL)
}

func TestCombineResultsInOrder_Empty(t *testing.T) {
	result := cache.CombineResultsInOrder(nil, nil, nil)
	assert.Len(t, result, 0)
}

// FIXED: Use require.Error instead of assert.Error for immediate failure
func TestHandleOutput_InvalidFormat(t *testing.T) {
	config := &types.AnalysisConfig{
		OutputFile:   "report.weird",
		OutputFormat: "weird",
	}
	results := []*types.PageResult{
		createMockResult("https://example.com", makeScores(1, 0.9, 1, 1), nil),
	}

	err := handleOutput(config, results, 1*time.Second)

	// Use require.Error to fail immediately if no error is returned
	require.Error(t, err, "Expected error for invalid output format 'weird'")
	assert.Contains(t, err.Error(), "failed to generate")
}

func TestHandleOutput_Stdout(t *testing.T) {
	config := &types.AnalysisConfig{
		UseStdout: true,
	}
	results := []*types.PageResult{
		createMockResult("https://example.com", makeScores(1, 0.9, 1, 1), nil),
	}
	err := handleOutput(config, results, 2*time.Second)
	assert.NoError(t, err)
}

func TestHandleOutput_StartServer(t *testing.T) {
	t.Skip("serverImpl.Start is not mocked — requires integration or test double")
}

func TestExecuteAnalysis_InvalidSitemap(t *testing.T) {
	config := &types.AnalysisConfig{
		Sitemap:      "invalid/path.xml",
		OutputFormat: "json",
		UseStdout:    true,
	}
	err := executeAnalysis(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse input")
}

func TestExecuteAnalysis_CacheOnly(t *testing.T) {
	t.Skip("Needs cache mock or pre-generated file to test fully")
}
