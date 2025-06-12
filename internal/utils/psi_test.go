package utils

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/stretchr/testify/assert"
)

// mockTransport implements http.RoundTripper for mocking HTTP responses
type mockTransport struct {
	req       *http.Request
	resp      *http.Response
	err       error
	urlCheck  func(*http.Request)
	callCount int
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.req = req
	t.callCount++
	if t.urlCheck != nil {
		t.urlCheck(req)
	}
	if t.err != nil {
		return nil, t.err
	}
	return t.resp, nil
}

// Test data helper
func createValidPSIResponse() string {
	return `{
		"lighthouseResult": {
			"categories": {
				"performance": {"id": "performance", "score": 0.9},
				"accessibility": {"id": "accessibility", "score": 0.85},
				"best-practices": {"id": "best-practices", "score": 0.95},
				"seo": {"id": "seo", "score": 0.88}
			},
			"audits": {
				"first-contentful-paint": {"numericValue": 1200},
				"largest-contentful-paint": {"numericValue": 2500},
				"cumulative-layout-shift": {"numericValue": 0.05},
				"unused-css-rules": {
					"title": "Remove unused CSS",
					"description": "Remove dead rules from stylesheets",
					"numericValue": 500,
					"score": 0.2,
					"details": {"key": "value"}
				}
			},
			"finalDisplayedUrl": "https://example.com/final",
			"userAgent": "test-agent"
		},
		"loadingExperience": {
			"originFallback": true,
			"metrics": {
				"FIRST_CONTENTFUL_PAINT_MS": {"percentile": 1100, "category": "FAST"}
			}
		}
	}`
}

func TestFetchScore_Success(t *testing.T) {
	mockResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(createValidPSIResponse())),
	}
	transport := &mockTransport{
		resp: mockResp,
		urlCheck: func(req *http.Request) {
			assert.Contains(t, req.URL.String(), "strategy=mobile")
			assert.Contains(t, req.URL.String(), "category=performance")
			assert.Contains(t, req.URL.String(), "url=https%3A%2F%2Fexample.com")
		},
	}
	httpClient = &http.Client{Transport: transport}

	result := FetchScore("https://example.com", "mobile")

	assert.NoError(t, result.Error)
	assert.Equal(t, "https://example.com", result.URL)
	assert.Equal(t, "mobile", result.Strategy)
	assert.Equal(t, "https://example.com/final", result.FinalURL)
	assert.True(t, result.Elapsed > 0)

	// Verify core scores are extracted
	assert.NotNil(t, result.Scores)
	assert.Equal(t, 0.9*constants.ScoreMultiplier, result.Scores.Performance)
	assert.Equal(t, 0.85*constants.ScoreMultiplier, result.Scores.Accessibility)

	// Verify core metrics are extracted
	assert.NotNil(t, result.Metrics)
	assert.Equal(t, 1200.0, result.Metrics.FirstContentfulPaint)
	assert.Equal(t, 2500.0, result.Metrics.LargestContentfulPaint)

	// Verify opportunities are found
	assert.NotNil(t, result.Opportunities)
	assert.Len(t, result.Opportunities, 1)
	assert.Equal(t, "unused-css-rules", result.Opportunities[0].ID)
}

func TestFetchScore_WithAPIKey(t *testing.T) {
	os.Setenv("PSI_API_KEY", "test-key")
	defer os.Unsetenv("PSI_API_KEY")

	mockResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"lighthouseResult": {"categories": {}}}`)),
	}
	transport := &mockTransport{
		resp: mockResp,
		urlCheck: func(req *http.Request) {
			assert.Contains(t, req.URL.String(), "key=test-key")
		},
	}
	httpClient = &http.Client{Transport: transport}

	result := FetchScore("https://example.com", "desktop")
	assert.NoError(t, result.Error)
}

func TestFetchScore_WithoutAPIKey(t *testing.T) {
	os.Unsetenv("PSI_API_KEY")

	mockResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"lighthouseResult": {"categories": {}}}`)),
	}
	transport := &mockTransport{
		resp: mockResp,
		urlCheck: func(req *http.Request) {
			assert.NotContains(t, req.URL.String(), "key=")
		},
	}
	httpClient = &http.Client{Transport: transport}

	result := FetchScore("https://example.com", "desktop")
	assert.NoError(t, result.Error)
}

func TestFetchScore_NetworkError(t *testing.T) {
	transport := &mockTransport{
		err: errors.New("network timeout"),
	}
	httpClient = &http.Client{Transport: transport}

	result := FetchScore("https://example.com", "mobile")
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "network timeout")
	assert.Equal(t, "https://example.com", result.URL)
	assert.True(t, result.Elapsed > 0)
}

func TestFetchScore_HTTPErrors(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		expected   string
	}{
		{"BadRequest", 400, "API error: status 400"},
		{"RateLimit", 429, "API error: status 429"},
		{"ServerError", 500, "API error: status 500"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockResp := &http.Response{
				StatusCode: tc.statusCode,
				Body:       io.NopCloser(strings.NewReader("")),
			}
			transport := &mockTransport{resp: mockResp}
			httpClient = &http.Client{Transport: transport}

			result := FetchScore("https://example.com", "mobile")
			assert.Error(t, result.Error)
			assert.Contains(t, result.Error.Error(), tc.expected)
		})
	}
}

func TestFetchScore_InvalidJSON(t *testing.T) {
	mockResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("invalid json")),
	}
	transport := &mockTransport{resp: mockResp}
	httpClient = &http.Client{Transport: transport}

	result := FetchScore("https://example.com", "mobile")
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "JSON parse error")
}

func TestFetchScore_EmptyResponse(t *testing.T) {
	mockResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}
	transport := &mockTransport{resp: mockResp}
	httpClient = &http.Client{Transport: transport}

	result := FetchScore("https://example.com", "mobile")
	assert.NoError(t, result.Error)
	assert.Equal(t, "https://example.com", result.URL)
	assert.Nil(t, result.Scores)
}

func TestFetchScore_PartialResponse(t *testing.T) {
	partialJSON := `{
		"lighthouseResult": {
			"categories": {
				"performance": {"score": 0.8}
			}
		}
	}`

	mockResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(partialJSON)),
	}
	transport := &mockTransport{resp: mockResp}
	httpClient = &http.Client{Transport: transport}

	result := FetchScore("https://example.com", "mobile")
	assert.NoError(t, result.Error)
	assert.NotNil(t, result.Scores)
	assert.Equal(t, 0.8*constants.ScoreMultiplier, result.Scores.Performance)
	assert.Equal(t, float64(0), result.Scores.Accessibility) // Missing categories default to 0
}
