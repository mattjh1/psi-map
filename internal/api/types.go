package api

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/types"
)

// APIPageResult represents page results for API responses (swagger-friendly)
type APIPageResult struct {
	URL      string     `json:"url" example:"https://example.com"`
	Mobile   *APIResult `json:"mobile,omitempty"`
	Desktop  *APIResult `json:"desktop,omitempty"`
	Duration int64      `json:"duration" example:"5000000000"` // nanoseconds
} // @name APIPageResult

// APIResult represents a page analysis result for API responses (swagger-friendly)
type APIResult struct {
	URL           string             `json:"url" example:"https://example.com"`
	FinalURL      string             `json:"final_url,omitempty" example:"https://example.com/"`
	Strategy      string             `json:"strategy" example:"mobile"`
	UserAgent     string             `json:"user_agent,omitempty" example:"Mozilla/5.0..."`
	Elapsed       int64              `json:"elapsed" example:"1500000000"` // nanoseconds
	Error         string             `json:"error,omitempty" example:""`
	Scores        *APICategoryScores `json:"scores,omitempty"`
	Metrics       *APIMetrics        `json:"metrics,omitempty"`
	FieldData     *APIFieldData      `json:"field_data,omitempty"`
	Opportunities []APIOpportunity   `json:"opportunities,omitempty"`
} // @name APIResult

// APICategoryScores holds scores for all Lighthouse categories
type APICategoryScores struct {
	Performance   float64 `json:"performance" example:"85.5"`
	Accessibility float64 `json:"accessibility" example:"92.1"`
	BestPractices float64 `json:"best_practices" example:"88.7"`
	SEO           float64 `json:"seo" example:"90.3"`
} // @name APICategoryScores

// APIMetrics contains core web vitals and performance metrics
type APIMetrics struct {
	// Core Web Vitals
	FirstContentfulPaint   float64 `json:"first_contentful_paint" example:"1800.5"`   // FCP in ms
	LargestContentfulPaint float64 `json:"largest_contentful_paint" example:"2400.2"` // LCP in ms
	FirstInputDelay        float64 `json:"first_input_delay" example:"50.1"`          // FID in ms
	CumulativeLayoutShift  float64 `json:"cumulative_layout_shift" example:"0.05"`    // CLS score
	// Additional Performance Metrics
	SpeedIndex        float64 `json:"speed_index" example:"3200.1"`         // Speed Index in ms
	TimeToInteractive float64 `json:"time_to_interactive" example:"4500.3"` // TTI in ms
	TotalBlockingTime float64 `json:"total_blocking_time" example:"200.5"`  // TBT in ms
	// Resource Metrics
	DOMSize       float64 `json:"dom_size" example:"1500"`         // Number of DOM elements
	ResourceCount int     `json:"resource_count" example:"75"`     // Total resources loaded
	TransferSize  int64   `json:"transfer_size" example:"2048576"` // Total bytes transferred
} // @name APIMetrics

// APIFieldData represents real user metrics from Chrome UX Report
type APIFieldData struct {
	OriginFallback bool                      `json:"originFallback" example:"false"`
	Metrics        map[string]APIFieldMetric `json:"metrics"`
} // @name APIFieldData

// APIFieldMetric represents a field metric from real users
type APIFieldMetric struct {
	Percentile float64 `json:"percentile" example:"75.5"`
	Category   string  `json:"category" example:"FAST"` // "FAST", "AVERAGE", "SLOW"
} // @name APIFieldMetric

// APIOpportunity represents a performance improvement opportunity
type APIOpportunity struct {
	ID               string  `json:"id" example:"unused-css-rules"`
	Title            string  `json:"title" example:"Remove unused CSS"`
	Description      string  `json:"description" example:"Reduce unused rules from stylesheets"`
	Impact           string  `json:"impact" example:"High"`            // "High", "Medium", "Low"
	PotentialSavings float64 `json:"potentialSavings" example:"500.2"` // Time savings in ms
} // @name APIOpportunity

// Conversion functions from internal types to API types

// ToAPIPageResult converts a types.PageResult to an APIPageResult.
func ToAPIPageResult(pr *types.PageResult) *APIPageResult {
	if pr == nil {
		return nil
	}

	return &APIPageResult{
		URL:      pr.URL,
		Duration: int64(pr.Duration), // time.Duration to nanoseconds
		Mobile:   ToAPIResult(pr.Mobile),
		Desktop:  ToAPIResult(pr.Desktop),
	}
}

// ToAPIResult converts a types.Result to an APIResult.
func ToAPIResult(r *types.Result) *APIResult {
	if r == nil {
		return nil
	}

	apiResult := &APIResult{
		URL:       r.URL,
		FinalURL:  r.FinalURL,
		Strategy:  r.Strategy,
		UserAgent: r.UserAgent,
		Elapsed:   int64(r.Elapsed),
	}

	if r.Error != nil {
		apiResult.Error = r.Error.Error()
	}

	apiResult.Scores = ToAPICategoryScores(r.Scores)
	apiResult.Metrics = ToAPIMetrics(r.Metrics)
	apiResult.FieldData = ToAPIFieldData(r.FieldData)

	if r.Opportunities != nil {
		apiResult.Opportunities = make([]APIOpportunity, len(r.Opportunities))
		for i, opp := range r.Opportunities {
			apiResult.Opportunities[i] = ToAPIOpportunity(&opp)
		}
	}

	return apiResult
}

// ToAPICategoryScores converts a types.CategoryScores to APICategoryScores.
func ToAPICategoryScores(cs *types.CategoryScores) *APICategoryScores {
	if cs == nil {
		return nil
	}
	return &APICategoryScores{
		Performance:   cs.Performance,
		Accessibility: cs.Accessibility,
		BestPractices: cs.BestPractices,
		SEO:           cs.SEO,
	}
}

// ToAPIMetrics converts a types.Metrics to APIMetrics.
func ToAPIMetrics(m *types.Metrics) *APIMetrics {
	if m == nil {
		return nil
	}
	return &APIMetrics{
		FirstContentfulPaint:   m.FirstContentfulPaint,
		LargestContentfulPaint: m.LargestContentfulPaint,
		FirstInputDelay:        m.FirstInputDelay,
		CumulativeLayoutShift:  m.CumulativeLayoutShift,
		SpeedIndex:             m.SpeedIndex,
		TimeToInteractive:      m.TimeToInteractive,
		TotalBlockingTime:      m.TotalBlockingTime,
		DOMSize:                m.DOMSize,
		ResourceCount:          m.ResourceCount,
		TransferSize:           m.TransferSize,
	}
}

// ToAPIFieldData converts a types.FieldData to APIFieldData.
func ToAPIFieldData(fd *types.FieldData) *APIFieldData {
	if fd == nil {
		return nil
	}

	apiMetrics := make(map[string]APIFieldMetric)
	for key, metric := range fd.Metrics {
		apiMetrics[key] = APIFieldMetric{
			Percentile: metric.Percentile,
			Category:   metric.Category,
		}
	}

	return &APIFieldData{
		OriginFallback: fd.OriginFallback,
		Metrics:        apiMetrics,
	}
}

// ToAPIOpportunity converts a types.Opportunity to an APIOpportunity.
func ToAPIOpportunity(o *types.Opportunity) APIOpportunity {
	if o == nil {
		return APIOpportunity{}
	}
	return APIOpportunity{
		ID:               o.ID,
		Title:            o.Title,
		Description:      o.Description,
		Impact:           o.Impact,
		PotentialSavings: o.PotentialSavings,
	}
}

// ConvertPageResults converts a slice of types.PageResult to []*APIPageResult.
func ConvertPageResults(internal []*types.PageResult) []*APIPageResult {
	if internal == nil {
		return nil
	}

	api := make([]*APIPageResult, len(internal))
	for i, pr := range internal {
		api[i] = ToAPIPageResult(pr)
	}
	return api
}

// Updated response types to use API-friendly types
// Replace your existing AnalysisResponse and JobStatusResponse with these:

// AnalysisResponse represents the response for completed analysis
type AnalysisResponse struct {
	Status    string           `json:"status" example:"completed"`
	Timestamp time.Time        `json:"timestamp"`
	Results   []*APIPageResult `json:"results"` // Changed to API type
	Summary   AnalysisSummary  `json:"summary"`
	Duration  string           `json:"duration,omitempty" example:"2m30s"`
} // @name AnalysisResponse

// JobStatusResponse represents the response for job status queries
type JobStatusResponse struct {
	JobID     string           `json:"job_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status    string           `json:"status" example:"running"`
	Progress  JobProgress      `json:"progress"`
	Timestamp time.Time        `json:"timestamp"`
	Results   []*APIPageResult `json:"results,omitempty"` // Changed to API type
	Summary   *AnalysisSummary `json:"summary,omitempty"`
	Duration  string           `json:"duration,omitempty" example:"45s"`
	Error     string           `json:"error,omitempty"`
} // @name JobStatusResponse

// AsyncAnalysisResponse represents the response for async analysis requests
type AsyncAnalysisResponse struct {
	JobID     string    `json:"job_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status    string    `json:"status" example:"started"`
	Timestamp time.Time `json:"timestamp"`
	StatusURL string    `json:"status_url" example:"/api/v1/analyze/status/550e8400-e29b-41d4-a716-446655440000"`
} // @name AsyncAnalysisResponse

// JobProgress represents the progress of an analysis job
type JobProgress struct {
	Current int `json:"current" example:"5"`
	Total   int `json:"total" example:"20"`
	Percent int `json:"percent" example:"25"`
} // @name JobProgress

// AnalysisSummary provides aggregate statistics
type AnalysisSummary struct {
	TotalURLs            int     `json:"total_urls" example:"20"`
	AnalyzedURLs         int     `json:"analyzed_urls" example:"18"`
	FailedURLs           int     `json:"failed_urls" example:"2"`
	AveragePerformance   float64 `json:"average_performance" example:"85.5"`
	AverageAccessibility float64 `json:"average_accessibility" example:"92.1"`
	AverageBestPractices float64 `json:"average_best_practices" example:"88.7"`
	AverageSEO           float64 `json:"average_seo" example:"90.3"`
	FastestURL           string  `json:"fastest_url,omitempty" example:"https://example.com/page1"`
	SlowestURL           string  `json:"slowest_url,omitempty" example:"https://example.com/page2"`
} // @name AnalysisSummary

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status" example:"healthy"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version" example:"v1.0.0"`
	Uptime    string    `json:"uptime" example:"2h30m15s"`
} // @name HealthResponse

// VersionResponse represents the version info response
type VersionResponse struct {
	Version   string `json:"version" example:"v1.0.0"`
	Commit    string `json:"commit" example:"abc123def"`
	BuildTime string `json:"build_time" example:"2025-07-30T10:00:00Z"`
} // @name VersionResponse

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string    `json:"error" example:"Invalid request"`
	Code    int       `json:"code" example:"400"`
	Message string    `json:"message" example:"The sitemap parameter is required"`
	Time    time.Time `json:"timestamp"`
} // @name ErrorResponse

// JobListResponse represents a list of jobs
type JobListResponse struct {
	Jobs  []JobSummary `json:"jobs"`
	Total int          `json:"total" example:"10"`
	Page  int          `json:"page" example:"1"`
	Limit int          `json:"limit" example:"50"`
} // @name JobListResponse

// JobSummary provides a summary of a job
type JobSummary struct {
	JobID     string      `json:"job_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status    string      `json:"status" example:"completed"`
	Created   time.Time   `json:"created"`
	Completed *time.Time  `json:"completed,omitempty"`
	Sitemap   string      `json:"sitemap" example:"https://example.com/sitemap.xml"`
	Progress  JobProgress `json:"progress"`
} // @name JobSummary

// AnalysisRequest represents the request payload for analysis with authentication support
type AnalysisRequest struct {
	Sitemap       string                `json:"sitemap" binding:"required" example:"https://example.com/sitemap.xml"`
	MaxWorkers    int                   `json:"max_workers,omitempty" example:"4"`
	CacheTTL      int                   `json:"cache_ttl,omitempty" example:"86400"`
	Provider      string                `json:"provider,omitempty" example:"psi"`
	LighthouseURL string                `json:"lighthouse_url,omitempty" example:"http://localhost:9222"`
	Async         bool                  `json:"async,omitempty" example:"true"`
	Auth          *AuthenticationConfig `json:"auth,omitempty"`
} // @name AnalysisRequest

// AuthenticationConfig holds authentication configuration for external services
type AuthenticationConfig struct {
	// Bearer token for Authorization header
	BearerToken string `json:"bearer_token,omitempty" example:"your-bearer-token"`

	// Cloudflare Access configuration
	CloudflareAccess *CloudflareAccessConfig `json:"cloudflare_access,omitempty"`
} // @name AuthenticationConfig

// CloudflareAccessConfig holds Cloudflare Access authentication details
type CloudflareAccessConfig struct {
	ClientID     string `json:"client_id,omitempty" example:"your-cf-client-id"`
	ClientSecret string `json:"client_secret,omitempty" example:"your-cf-client-secret"`
} // @name CloudflareAccessConfig

// Validate validates the analysis request
func (r *AnalysisRequest) Validate() error {
	var lh = "lighthouse"

	if strings.TrimSpace(r.Sitemap) == "" {
		return fmt.Errorf("sitemap is required")
	}

	if r.MaxWorkers < 0 {
		return fmt.Errorf("max_workers must be >= 0")
	}

	if r.CacheTTL < 0 {
		return fmt.Errorf("cache_ttl must be >= 0")
	}

	if r.Provider != "" && r.Provider != "psi" && r.Provider != lh {
		return fmt.Errorf("provider must be 'psi' or 'lighthouse'")
	}

	if r.Provider == lh && strings.TrimSpace(r.LighthouseURL) == "" {
		return fmt.Errorf("lighthouse_url is required when provider is 'lighthouse'")
	}

	// Validate authentication configuration if using lighthouse provider
	if r.Provider == lh && r.Auth != nil {
		if r.Auth.CloudflareAccess != nil {
			if r.Auth.CloudflareAccess.ClientID == "" || r.Auth.CloudflareAccess.ClientSecret == "" {
				return fmt.Errorf("both client_id and client_secret are required for Cloudflare Access")
			}
		}
	}

	return nil
}

// SetDefaults sets default values for optional fields
func (r *AnalysisRequest) SetDefaults() {
	if r.MaxWorkers == 0 {
		r.MaxWorkers = max(1, runtime.NumCPU()/constants.CPUDivisor)
	}
	if r.CacheTTL == 0 {
		r.CacheTTL = constants.DefaultTTLHours
	}
	if r.Provider == "" {
		r.Provider = "psi"
	}
}
