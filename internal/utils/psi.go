package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/types/psi"
)

// Package-level HTTP client for testability
var httpClient = &http.Client{}

// FetchScore retrieves comprehensive performance data from the PSI API
func FetchScoreImpl(ctx context.Context, pageURL, strategy string) types.Result {
	start := time.Now()
	apiKey := os.Getenv("PSI_API_KEY")
	baseURL := "https://www.googleapis.com/pagespeedonline/v5/runPagespeed"
	params := url.Values{}
	params.Add("url", pageURL)
	params.Add("strategy", strategy)
	params.Add("category", "performance")
	params.Add("category", "accessibility")
	params.Add("category", "best-practices")
	params.Add("category", "seo")
	if apiKey != "" {
		params.Add("key", apiKey)
	}
	fullURL := baseURL + "?" + params.Encode()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, http.NoBody)
	if err != nil {
		return types.Result{
			URL:      pageURL,
			Strategy: strategy,
			Error:    fmt.Errorf("failed to create request: %w", err),
			Elapsed:  time.Since(start),
		}
	}

	// Use httpClient.Do instead of httpClient.Get
	resp, err := httpClient.Do(req)
	if err != nil {
		return types.Result{
			URL:      pageURL,
			Strategy: strategy,
			Error:    fmt.Errorf("request failed: %w", err),
			Elapsed:  time.Since(start),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.Result{
			URL:      pageURL,
			Strategy: strategy,
			Error:    fmt.Errorf("API error: status %d", resp.StatusCode),
			Elapsed:  time.Since(start),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.Result{
			URL:      pageURL,
			Strategy: strategy,
			Error:    fmt.Errorf("read body error: %w", err),
			Elapsed:  time.Since(start),
		}
	}

	var data psi.PSIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return types.Result{
			URL:      pageURL,
			Strategy: strategy,
			Error:    fmt.Errorf("JSON parse error: %w", err),
			Elapsed:  time.Since(start),
		}
	}

	return extractResultData(&data, pageURL, strategy, time.Since(start))
}

// If you want to provide a convenience function without context for backward compatibility:
func FetchScoreWithTimeout(pageURL, strategy string, timeout time.Duration) types.Result {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return FetchScoreImpl(ctx, pageURL, strategy)
}

func FetchScore(pageURL, strategy string) types.Result {
	return FetchScoreWithTimeout(pageURL, strategy, constants.ReadHeaderTimeout)
}

// extractResultData processes the PSI response into our Result struct
func extractResultData(data *psi.PSIResponse, pageURL, strategy string, elapsed time.Duration) types.Result {
	result := types.Result{
		URL:      pageURL,
		Strategy: strategy,
		Elapsed:  elapsed,
	}

	// Extract category scores
	if lr := data.LighthouseResult; lr != nil {
		if lr.Categories != nil {
			result.Scores = &types.CategoryScores{
				Performance:   getScore(lr.Categories.Performance),
				Accessibility: getScore(lr.Categories.Accessibility),
				BestPractices: getScore(lr.Categories.BestPractices),
				SEO:           getScore(lr.Categories.SEO),
			}
		}

		if lr.Audits != nil {
			result.Metrics = extractMetrics(lr.Audits)
			result.Opportunities = extractOpportunities(lr.Audits)
		}

		if lr.FinalDisplayedURL != "" {
			result.FinalURL = lr.FinalDisplayedURL
		}

		result.UserAgent = lr.UserAgent
	}

	// Extract loading experience data
	if data.LoadingExperience != nil {
		result.FieldData = extractFieldData(data.LoadingExperience)
	}

	return result
}

// getScore safely extracts score from category, handling nil cases
func getScore(category *psi.Category) float64 {
	if category == nil || category.Score == nil {
		return 0
	}
	return *category.Score * constants.ScoreMultiplier
}

// extractMetrics pulls out the key performance metrics
func extractMetrics(audits map[string]*psi.Audit) *types.Metrics {
	if audits == nil {
		return nil
	}

	return &types.Metrics{
		FirstContentfulPaint:   getMetricValue(audits["first-contentful-paint"]),
		LargestContentfulPaint: getMetricValue(audits["largest-contentful-paint"]),
		FirstInputDelay:        getMetricValue(audits["max-potential-fid"]),
		CumulativeLayoutShift:  getNumericValue(audits["cumulative-layout-shift"]),
		SpeedIndex:             getMetricValue(audits["speed-index"]),
		TimeToInteractive:      getMetricValue(audits["interactive"]),
		TotalBlockingTime:      getMetricValue(audits["total-blocking-time"]),

		// Resource metrics
		DOMSize:       getNumericValue(audits["dom-size"]),
		ResourceCount: getResourceCount(audits),
		TransferSize:  getTotalTransferSize(audits),
	}
}

// extractOpportunities finds the biggest performance improvement opportunities
func extractOpportunities(audits map[string]*psi.Audit) []types.Opportunity {
	if audits == nil {
		return nil
	}

	var opportunities []types.Opportunity

	// Key opportunity audits to check
	opportunityAudits := []string{
		"unused-css-rules",
		"unused-javascript",
		"modern-image-formats",
		"efficiently-encode-images",
		"render-blocking-resources",
		"unminified-css",
		"unminified-javascript",
		"legacy-javascript",
		"largest-contentful-paint-element",
	}

	for _, auditKey := range opportunityAudits {
		if audit := audits[auditKey]; audit != nil && audit.Details != nil {
			opp := types.Opportunity{
				ID:          auditKey,
				Title:       audit.Title,
				Description: audit.Description,
			}

			// Extract potential savings
			if audit.NumericValue != nil {
				opp.PotentialSavings = *audit.NumericValue
			}

			// Get impact level based on score
			if audit.Score != nil {
				switch {
				case *audit.Score < constants.AuditScorePoorThreshold:
					opp.Impact = "High"
				case *audit.Score < constants.AuditScoreGoodThreshold:
					opp.Impact = "Medium"
				default:
					opp.Impact = "Low"
				}
			}

			opportunities = append(opportunities, opp)
		}
	}

	return opportunities
}

// extractFieldData processes real user metrics if available
func extractFieldData(loadingExp *psi.LoadingExperience) *types.FieldData {
	if loadingExp == nil {
		return nil
	}

	fieldData := &types.FieldData{
		OriginFallback: loadingExp.OriginFallback,
	}

	// Extract field metrics
	if loadingExp.Metrics != nil {
		fieldData.Metrics = make(map[string]types.FieldMetric)

		for key, metric := range loadingExp.Metrics {
			if metric != nil {
				fieldData.Metrics[key] = types.FieldMetric{
					Percentile: metric.Percentile,
					Category:   metric.Category,
				}
			}
		}
	}

	return fieldData
}

// Helper functions for extracting specific metric values
func getMetricValue(audit *psi.Audit) float64 {
	if audit == nil || audit.NumericValue == nil {
		return 0
	}
	return *audit.NumericValue
}

func getNumericValue(audit *psi.Audit) float64 {
	if audit == nil || audit.NumericValue == nil {
		return 0
	}
	return *audit.NumericValue
}

func getResourceCount(audits map[string]*psi.Audit) int {
	// Sum up various resource counts
	count := 0
	resourceAudits := []string{
		"network-requests",
		"resource-summary",
	}

	for _, auditKey := range resourceAudits {
		if audit := audits[auditKey]; audit != nil && audit.Details != nil {
			// TODO: Implement parsing for audit.Details
			_ = audit // silence unused warning
		}
	}

	return count
}

//lint:ignore unparam placeholder for future implementation
func getTotalTransferSize(audits map[string]*psi.Audit) int64 {
	if audit := audits["network-requests"]; audit != nil && audit.Details != nil {
		// Parse network requests to sum transfer sizes
		// This would need detailed parsing of the audit.Details structure
		_ = audit
	}
	return 0
}
