package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/mattjh1/psi-map/internal/logger"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/stretchr/testify/assert"
)

// Mock server.GenerateHTMLFile
var mockGenerateHTMLFile func(results []*types.PageResult, filename string) error

func init() {
	// Replace the actual function with a mock for testing
	serverGenerateHTMLFile = func(results []*types.PageResult, filename string) error {
		if mockGenerateHTMLFile != nil {
			return mockGenerateHTMLFile(results, filename)
		}
		return fmt.Errorf("mock not set")
	}
}

func TestSaveJSONReport(t *testing.T) {
	results := []*types.PageResult{
		{URL: "http://example.com/page1"},
		{URL: "http://example.com/page2"},
	}
	filename := "test_report.json"
	defer os.Remove(filename)

	err := SaveJSONReport(results, filename)
	assert.NoError(t, err)

	content, err := os.ReadFile(filename)
	assert.NoError(t, err)

	var decodedResults []*types.PageResult
	err = json.Unmarshal(content, &decodedResults)
	assert.NoError(t, err)
	assert.Equal(t, results, decodedResults)
}

func TestSaveJSONToStdout(t *testing.T) {
	results := []*types.PageResult{
		{URL: "http://example.com/stdout1"},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := SaveJSONToStdout(results)
	assert.NoError(t, err)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	var decodedResults []*types.PageResult
	err = json.Unmarshal(out, &decodedResults)
	assert.NoError(t, err)
	assert.Equal(t, results, decodedResults)
}

func TestSaveHTMLReport(t *testing.T) {
	results := []*types.PageResult{{URL: "http://example.com/html"}}
	filename := "test_report.html"

	// Set the mock to return no error
	mockGenerateHTMLFile = func(r []*types.PageResult, f string) error {
		assert.Equal(t, results, r)
		assert.Equal(t, filename, f)
		return nil
	}

	err := SaveHTMLReport(results, filename)
	assert.NoError(t, err)

	// Set the mock to return an error
	mockGenerateHTMLFile = func(r []*types.PageResult, f string) error {
		return fmt.Errorf("mock HTML generation error")
	}

	err = SaveHTMLReport(results, filename)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate HTML report")
}

func TestFormatCategoryName(t *testing.T) {
	assert.Equal(t, "Performance", formatCategoryName("performance"))
	assert.Equal(t, "Accessibility", formatCategoryName("accessibility"))
	assert.Equal(t, "Best Practices", formatCategoryName("best_practices"))
	assert.Equal(t, "SEO", formatCategoryName("seo"))
	assert.Equal(t, "unknown_category", formatCategoryName("unknown_category"))
}

func TestPrintSummary(t *testing.T) {
	logger.Reset()

	// Mock the logger's output
	var buf bytes.Buffer
	logger.Init(logger.WithOutput(&buf)) // Ensure singleton is also mocked

	results := []*types.PageResult{
		{URL: "http://example.com/good", Mobile: &types.Result{Scores: &types.CategoryScores{Performance: 95, Accessibility: 90, BestPractices: 85, SEO: 80}}},
		{URL: "http://example.com/needs_improvement", Mobile: &types.Result{Scores: &types.CategoryScores{Performance: 60, Accessibility: 65, BestPractices: 70, SEO: 75}}},
		{URL: "http://example.com/poor", Mobile: &types.Result{Scores: &types.CategoryScores{Performance: 30, Accessibility: 35, BestPractices: 40, SEO: 45}}},
		{URL: "http://example.com/failed", Mobile: &types.Result{Error: fmt.Errorf("failed to analyze")}, Desktop: &types.Result{Error: fmt.Errorf("failed to analyze")}},
	}
	elapsed := 10 * time.Second

	PrintSummary(results, elapsed)

	output := buf.String()
	// Remove ANSI escape codes for reliable assertion
	output = removeANSI(output)

	assert.Contains(t, output, "SUMMARY")
	assert.Contains(t, output, "Total Pages Analyzed: 4")
	assert.Contains(t, output, "Successful: 3")
	assert.Contains(t, output, "Failed: 1")
	assert.Contains(t, output, "Average Scores")
	assert.Contains(t, output, "Performance: 61.7")
	assert.Contains(t, output, "Accessibility: 63.3")
	assert.Contains(t, output, "Best Practices: 65.0")
	assert.Contains(t, output, "SEO: 66.7")
	assert.Contains(t, output, "Score Distribution")
	assert.Contains(t, output, "Performance: Good: 1, Needs Improvement: 1, Poor: 1")
	assert.Contains(t, output, "Accessibility: Good: 1, Needs Improvement: 1, Poor: 1")
	assert.Contains(t, output, "Best Practices: Good: 0, Needs Improvement: 2, Poor: 1")
	assert.Contains(t, output, "SEO: Good: 0, Needs Improvement: 2, Poor: 1")
	assert.Contains(t, output, "Total Time Elapsed: 10s")
}

// removeANSI removes ANSI escape codes from a string
func removeANSI(s string) string {
	const ansi = "\u001b[[0-9;]*m"
	re := regexp.MustCompile(ansi)
	return re.ReplaceAllString(s, "")
}
