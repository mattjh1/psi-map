package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/mattjh1/psi-map/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestCalculateSitemapHash(t *testing.T) {
	// Test with URLs only
	hash1, err := calculateSitemapHash("", []string{"http://example.com/1", "http://example.com/2"})
	assert.NoError(t, err)
	assert.NotEmpty(t, hash1)

	hash2, err := calculateSitemapHash("", []string{"http://example.com/1", "http://example.com/2"})
	assert.NoError(t, err)
	assert.Equal(t, hash1, hash2, "Hashes for same URLs should be identical")

	hash3, err := calculateSitemapHash("", []string{"http://example.com/3"})
	assert.NoError(t, err)
	assert.NotEqual(t, hash1, hash3, "Hashes for different URLs should be different")

	// Test with sitemap path (mock file system if needed, for now, just ensure it doesn't error)
	// This part would ideally involve mocking os.Open and io.Copy
	// For now, we'll just test a non-existent path to ensure error handling (or lack thereof) is consistent
	_, err = calculateSitemapHash("non_existent_sitemap.xml", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open sitemap")
}

func TestGetURLCacheFilename(t *testing.T) {
	cacheDir := "/tmp/test_cache"
	url := "http://example.com/page"

	filename := getURLCacheFilename(cacheDir, url)
	assert.Equal(t, "/tmp/test_cache/urls/url-c7208ac94afcd66b5d5cd1dc5fc49c8b.json", filename)
}

func TestGetSitemapIndexFilename(t *testing.T) {
	cacheDir := "/tmp/test_cache"
	sitemapHash := "abcdef123456"

	filename := getSitemapIndexFilename(cacheDir, sitemapHash)
	assert.Equal(t, "/tmp/test_cache/indexes/sitemap-abcdef123456.json", filename)
}

func TestExtractPerformanceScore(t *testing.T) {
	// Test with mobile score
	result1 := &types.PageResult{
		Mobile: &types.Result{Scores: &types.CategoryScores{Performance: 90.5}},
	}
	assert.Equal(t, 90.5, extractPerformanceScore(result1))

	// Test with desktop score (mobile is 0)
	result2 := &types.PageResult{
		Mobile:  &types.Result{Scores: &types.CategoryScores{Performance: 0}},
		Desktop: &types.Result{Scores: &types.CategoryScores{Performance: 85.0}},
	}
	assert.Equal(t, 85.0, extractPerformanceScore(result2))

	// Test with no scores
	result3 := &types.PageResult{}
	assert.Equal(t, 0.0, extractPerformanceScore(result3))

	// Test with both scores, mobile should be preferred
	result4 := &types.PageResult{
		Mobile:  &types.Result{Scores: &types.CategoryScores{Performance: 95.0}},
		Desktop: &types.Result{Scores: &types.CategoryScores{Performance: 80.0}},
	}
	assert.Equal(t, 95.0, extractPerformanceScore(result4))
}

func TestHasErrors(t *testing.T) {
	// Test with mobile error
	result1 := &types.PageResult{
		Mobile: &types.Result{Error: fmt.Errorf("mobile error")},
	}
	assert.True(t, hasErrors(result1))

	// Test with desktop error
	result2 := &types.PageResult{
		Desktop: &types.Result{Error: fmt.Errorf("desktop error")},
	}
	assert.True(t, hasErrors(result2))

	// Test with no scores (implies error)
	result3 := &types.PageResult{}
	assert.True(t, hasErrors(result3))

	// Test with low mobile performance score
	result4 := &types.PageResult{
		Mobile: &types.Result{Scores: &types.CategoryScores{Performance: 49}},
	}
	assert.True(t, hasErrors(result4))

	// Test with low desktop performance score
	result5 := &types.PageResult{
		Desktop: &types.Result{Scores: &types.CategoryScores{Performance: 49}},
	}
	assert.True(t, hasErrors(result5))

	// Test with no errors and good scores
	result6 := &types.PageResult{
		Mobile:  &types.Result{Scores: &types.CategoryScores{Performance: 90}},
		Desktop: &types.Result{Scores: &types.CategoryScores{Performance: 80}},
	}
	assert.False(t, hasErrors(result6))
}

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "30m", formatDuration(30*time.Minute))
	assert.Equal(t, "1.5h", formatDuration(90*time.Minute))
	assert.Equal(t, "1.0d", formatDuration(24*time.Hour))
	assert.Equal(t, "2.5d", formatDuration(60*time.Hour))
}
