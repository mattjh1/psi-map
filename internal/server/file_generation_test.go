package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattjh1/psi-map/internal/types"
)

func TestGenerateHTMLFile(t *testing.T) {
	results := []types.PageResult{
		createMockResult("https://example.com", 90, 85, 80, 95, false),
	}

	// Create temporary file
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test-report.html")

	err := GenerateHTMLFile(results, filename)
	require.NoError(t, err)

	// Check file was created
	_, err = os.Stat(filename)
	assert.NoError(t, err)

	// Check file has content
	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Contains(t, string(content), "https://example.com")
	assert.Contains(t, string(content), "<!DOCTYPE html>") // Basic HTML structure

	// Verify it contains expected score data
	assert.Contains(t, string(content), "90") // Performance score
	assert.Contains(t, string(content), "85") // Accessibility score
}

func TestGenerateHTMLFile_WithMultipleResults(t *testing.T) {
	results := []types.PageResult{
		createMockResult("https://example1.com", 90, 85, 80, 95, false),
		createMockResult("https://example2.com", 70, 75, 65, 80, false),
		createResultWithMetrics("https://example3.com", 85),
	}

	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "multi-report.html")

	err := GenerateHTMLFile(results, filename)
	require.NoError(t, err)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	// Check all URLs are present
	assert.Contains(t, string(content), "https://example1.com")
	assert.Contains(t, string(content), "https://example2.com")
	assert.Contains(t, string(content), "https://example3.com")

	// Check it contains HTML structure
	assert.Contains(t, string(content), "<!DOCTYPE html>")
	assert.Contains(t, string(content), "<html")
	assert.Contains(t, string(content), "</html>")
}

func TestGenerateHTMLFile_WithEmptyResults(t *testing.T) {
	results := []types.PageResult{}

	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "empty-report.html")

	err := GenerateHTMLFile(results, filename)
	require.NoError(t, err)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	// Should still generate valid HTML
	assert.Contains(t, string(content), "<!DOCTYPE html>")
	assert.Contains(t, string(content), "<html")
	assert.Contains(t, string(content), "</html>")
}

func TestGenerateHTMLFile_InvalidPath(t *testing.T) {
	results := []types.PageResult{
		createMockResult("https://example.com", 90, 85, 80, 95, false),
	}

	// Try to write to a non-existent directory without creating parent dirs
	invalidPath := "/nonexistent/directory/report.html"

	err := GenerateHTMLFile(results, invalidPath)
	assert.Error(t, err)
}

func TestGenerateHTMLFile_WithErrorResults(t *testing.T) {
	results := []types.PageResult{
		createMockResult("https://success.com", 90, 85, 80, 95, false),
		createMockResult("https://error.com", 0, 0, 0, 0, true),
		createResultWithPartialSuccess("https://partial.com", true, false),
	}

	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "error-report.html")

	err := GenerateHTMLFile(results, filename)
	require.NoError(t, err)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	// Should handle all result types
	assert.Contains(t, string(content), "https://success.com")
	assert.Contains(t, string(content), "https://error.com")
	assert.Contains(t, string(content), "https://partial.com")
}
