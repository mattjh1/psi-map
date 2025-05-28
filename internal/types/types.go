package types

import (
	"time"
)

// Sitemap represents the structure of a basic XML sitemap
type Sitemap struct {
	URLs []URL `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
}

type PageResult struct {
	URL      string
	Mobile   Result
	Desktop  Result
	Duration time.Duration
}

// Result represents the comprehensive PSI analysis result for a single URL
type Result struct {
	URL       string        `json:"url"`
	FinalURL  string        `json:"final_url,omitempty"`
	Strategy  string        `json:"strategy"`
	UserAgent string        `json:"user_agent,omitempty"`
	Elapsed   time.Duration `json:"elapsed"`
	Error     error         `json:"error,omitempty"`

	// Lighthouse scores for all categories
	Scores *CategoryScores `json:"scores,omitempty"`

	// Core Web Vitals and performance metrics
	Metrics *Metrics `json:"metrics,omitempty"`

	// Real user field data (when available)
	FieldData *FieldData `json:"field_data,omitempty"`

	// Performance improvement opportunities
	Opportunities []Opportunity `json:"opportunities,omitempty"`
}

// ReportSummary contains aggregate statistics
type ReportSummary struct {
	TotalPages        int                `json:"total_pages"`
	SuccessfulPages   int                `json:"successful_pages"`
	FailedPages       int                `json:"failed_pages"`
	AverageScores     map[string]float64 `json:"average_scores"`
	ScoreDistribution map[string][3]int  `json:"score_distribution"` // [good, needs_improvement, poor]
}

// ReportData represents the complete report structure
type ReportData struct {
	Generated time.Time     `json:"generated"`
	Summary   ReportSummary `json:"summary"`
	Results   []Result      `json:"results"`
}

// CategoryScores holds scores for all Lighthouse categories
type CategoryScores struct {
	Performance   float64 `json:"performance"`
	Accessibility float64 `json:"accessibility"`
	BestPractices float64 `json:"best_practices"`
	SEO           float64 `json:"seo"`
}

// Metrics contains core web vitals and performance metrics
type Metrics struct {
	// Core Web Vitals
	FirstContentfulPaint   float64 `json:"first_contentful_paint"`   // FCP in ms
	LargestContentfulPaint float64 `json:"largest_contentful_paint"` // LCP in ms
	FirstInputDelay        float64 `json:"first_input_delay"`        // FID in ms
	CumulativeLayoutShift  float64 `json:"cumulative_layout_shift"`  // CLS score

	// Additional Performance Metrics
	SpeedIndex        float64 `json:"speed_index"`         // Speed Index in ms
	TimeToInteractive float64 `json:"time_to_interactive"` // TTI in ms
	TotalBlockingTime float64 `json:"total_blocking_time"` // TBT in ms

	// Resource Metrics
	DOMSize       float64 `json:"dom_size"`       // Number of DOM elements
	ResourceCount int     `json:"resource_count"` // Total resources loaded
	TransferSize  int64   `json:"transfer_size"`  // Total bytes transferred
}

// GetCoreWebVitalsGrade returns letter grades for Core Web Vitals
func (m *Metrics) GetCoreWebVitalsGrade() map[string]string {
	grades := make(map[string]string)

	// FCP grading (< 1.8s = good, < 3s = needs improvement, >= 3s = poor)
	switch {
	case m.FirstContentfulPaint < 1800:
		grades["fcp"] = "good"
	case m.FirstContentfulPaint < 3000:
		grades["fcp"] = "needs-improvement"
	default:
		grades["fcp"] = "poor"
	}

	// LCP grading (< 2.5s = good, < 4s = needs improvement, >= 4s = poor)
	switch {
	case m.LargestContentfulPaint < 2500:
		grades["lcp"] = "good"
	case m.LargestContentfulPaint < 4000:
		grades["lcp"] = "needs-improvement"
	default:
		grades["lcp"] = "poor"
	}

	// CLS grading (< 0.1 = good, < 0.25 = needs improvement, >= 0.25 = poor)
	switch {
	case m.CumulativeLayoutShift < 0.1:
		grades["cls"] = "good"
	case m.CumulativeLayoutShift < 0.25:
		grades["cls"] = "needs-improvement"
	default:
		grades["cls"] = "poor"
	}

	// FID grading (< 100ms = good, < 300ms = needs improvement, >= 300ms = poor)
	switch {
	case m.FirstInputDelay < 100:
		grades["fid"] = "good"
	case m.FirstInputDelay < 300:
		grades["fid"] = "needs-improvement"
	default:
		grades["fid"] = "poor"
	}

	return grades
}

// FieldData represents real user metrics from Chrome UX Report
type FieldData struct {
	OriginFallback bool                   `json:"origin_fallback"`
	Metrics        map[string]FieldMetric `json:"metrics"`
}

// FieldMetric represents a field metric from real users
type FieldMetric struct {
	Percentile float64 `json:"percentile"`
	Category   string  `json:"category"` // "FAST", "AVERAGE", "SLOW"
}

// Opportunity represents a performance improvement opportunity
type Opportunity struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	Impact           string  `json:"impact"`            // "High", "Medium", "Low"
	PotentialSavings float64 `json:"potential_savings"` // Time savings in ms
}

// =================== PSI API Response Structures ===================

// PSIResponse represents the full PageSpeed Insights API response
type PSIResponse struct {
	CaptchaResult           string             `json:"captchaResult,omitempty"`
	Kind                    string             `json:"kind,omitempty"`
	ID                      string             `json:"id,omitempty"`
	LoadingExperience       *LoadingExperience `json:"loadingExperience,omitempty"`
	OriginLoadingExperience *LoadingExperience `json:"originLoadingExperience,omitempty"`
	LighthouseResult        *LighthouseResult  `json:"lighthouseResult,omitempty"`
	AnalysisUTCTimestamp    string             `json:"analysisUTCTimestamp,omitempty"`
}

// LighthouseResult contains the Lighthouse audit results
type LighthouseResult struct {
	RequestedUrl       string                    `json:"requestedUrl,omitempty"`
	FinalDisplayedURL  string                    `json:"finalDisplayedUrl,omitempty"`
	MainDocumentURL    string                    `json:"mainDocumentUrl,omitempty"`
	FinalUrl           string                    `json:"finalUrl,omitempty"`
	LighthouseVersion  string                    `json:"lighthouseVersion,omitempty"`
	UserAgent          string                    `json:"userAgent,omitempty"`
	FetchTime          string                    `json:"fetchTime,omitempty"`
	Environment        *Environment              `json:"environment,omitempty"`
	RunWarnings        []string                  `json:"runWarnings,omitempty"`
	ConfigSettings     *ConfigSettings           `json:"configSettings,omitempty"`
	Categories         *Categories               `json:"categories,omitempty"`
	CategoryGroups     map[string]*CategoryGroup `json:"categoryGroups,omitempty"`
	Audits             map[string]*Audit         `json:"audits,omitempty"`
	Timing             *Timing                   `json:"timing,omitempty"`
	I18n               *I18n                     `json:"i18n,omitempty"`
	Entities           []Entity                  `json:"entities,omitempty"`
	FullPageScreenshot *FullPageScreenshot       `json:"fullPageScreenshot,omitempty"`
}

// Categories contains all Lighthouse category results
type Categories struct {
	Performance   *Category `json:"performance,omitempty"`
	Accessibility *Category `json:"accessibility,omitempty"`
	BestPractices *Category `json:"best-practices,omitempty"`
	SEO           *Category `json:"seo,omitempty"`
	PWA           *Category `json:"pwa,omitempty"`
}

// Category represents a Lighthouse category (Performance, Accessibility, etc.)
type Category struct {
	ID                string     `json:"id,omitempty"`
	Title             string     `json:"title,omitempty"`
	Description       string     `json:"description,omitempty"`
	Score             *float64   `json:"score,omitempty"`
	ManualDescription string     `json:"manualDescription,omitempty"`
	AuditRefs         []AuditRef `json:"auditRefs,omitempty"`
}

// AuditRef represents a reference to an audit within a category
type AuditRef struct {
	ID             string   `json:"id"`
	Weight         float64  `json:"weight,omitempty"`
	Group          string   `json:"group,omitempty"`
	Acronym        string   `json:"acronym,omitempty"`
	RelevantAudits []string `json:"relevantAudits,omitempty"`
}

// Audit represents a single Lighthouse audit
type Audit struct {
	ID               string         `json:"id,omitempty"`
	Title            string         `json:"title,omitempty"`
	Description      string         `json:"description,omitempty"`
	Score            *float64       `json:"score,omitempty"`
	ScoreDisplayMode string         `json:"scoreDisplayMode,omitempty"`
	NumericValue     *float64       `json:"numericValue,omitempty"`
	NumericUnit      string         `json:"numericUnit,omitempty"`
	DisplayValue     string         `json:"displayValue,omitempty"`
	Explanation      string         `json:"explanation,omitempty"`
	ErrorMessage     string         `json:"errorMessage,omitempty"`
	Warnings         []string       `json:"warnings,omitempty"`
	Details          map[string]any `json:"details,omitempty"`
	MetricSavings    *MetricSavings `json:"metricSavings,omitempty"`
}

// MetricSavings represents potential savings from fixing an audit
type MetricSavings struct {
	FCP *float64 `json:"FCP,omitempty"`
	LCP *float64 `json:"LCP,omitempty"`
}

// LoadingExperience represents Chrome UX Report data
type LoadingExperience struct {
	ID              string                    `json:"id,omitempty"`
	Metrics         map[string]*LoadingMetric `json:"metrics,omitempty"`
	OverallCategory string                    `json:"overall_category,omitempty"`
	InitialUrl      string                    `json:"initial_url,omitempty"`
	OriginFallback  bool                      `json:"origin_fallback,omitempty"`
}

// LoadingMetric represents a single loading metric from field data
type LoadingMetric struct {
	Percentile    float64             `json:"percentile,omitempty"`
	Distributions []DistributionEntry `json:"distributions,omitempty"`
	Category      string              `json:"category,omitempty"`
}

// DistributionEntry represents the distribution of a metric
type DistributionEntry struct {
	Min        float64 `json:"min,omitempty"`
	Max        float64 `json:"max,omitempty"`
	Proportion float64 `json:"proportion,omitempty"`
}

// Environment contains information about the test environment
type Environment struct {
	NetworkUserAgent string  `json:"networkUserAgent,omitempty"`
	HostUserAgent    string  `json:"hostUserAgent,omitempty"`
	BenchmarkIndex   float64 `json:"benchmarkIndex,omitempty"`
}

// ConfigSettings contains Lighthouse configuration
type ConfigSettings struct {
	EmulatedFormFactor string   `json:"emulatedFormFactor,omitempty"`
	FormFactor         string   `json:"formFactor,omitempty"`
	Locale             string   `json:"locale,omitempty"`
	OnlyCategories     []string `json:"onlyCategories,omitempty"`
	Channel            string   `json:"channel,omitempty"`
}

// CategoryGroup represents a group of related audits
type CategoryGroup struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// Timing contains performance timing information
type Timing struct {
	Total float64 `json:"total,omitempty"`
}

// I18n contains internationalization data
type I18n struct {
	RendererFormattedStrings map[string]string `json:"rendererFormattedStrings,omitempty"`
}

// Entity represents a web entity (like a company or organization)
type Entity struct {
	Name         string   `json:"name,omitempty"`
	Origins      []string `json:"origins,omitempty"`
	Homepage     string   `json:"homepage,omitempty"`
	Categories   []string `json:"categories,omitempty"`
	IsFirstParty bool     `json:"isFirstParty,omitempty"`
}

// FullPageScreenshot contains screenshot data
type FullPageScreenshot struct {
	Screenshot *Screenshot      `json:"screenshot,omitempty"`
	Nodes      map[string]*Node `json:"nodes,omitempty"`
}

// Screenshot represents screenshot data
type Screenshot struct {
	Data   string `json:"data,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// Node represents a DOM node in the screenshot
type Node struct {
	Left   float64 `json:"left,omitempty"`
	Top    float64 `json:"top,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

// =================== Legacy compatibility ===================

// Lighthouse maintains backward compatibility with existing code
type Lighthouse struct {
	LighthouseResult struct {
		Categories struct {
			Performance struct {
				Score float64 `json:"score"`
			} `json:"performance"`
		} `json:"categories"`
	} `json:"lighthouseResult"`
}
