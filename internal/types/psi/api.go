// Package psi contains types for Google PageSpeed Insights API responses
package psi

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

// LoadingExperience represents Chrome UX Report data
type LoadingExperience struct {
	ID              string                    `json:"id,omitempty"`
	Metrics         map[string]*LoadingMetric `json:"metrics,omitempty"`
	OverallCategory string                    `json:"overall_category,omitempty"`
	InitialURL      string                    `json:"initial_url,omitempty"`
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
