package types

import (
	"time"
)

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
	TotalPages        int
	SuccessfulPages   int
	FailedPages       int
	AverageScores     map[string]float64
	ScoreDistribution map[string][]int // good, needs-improvement, poor counts
	FastestPage       *PageResult
	SlowestPage       *PageResult
	BestPerformance   *PageResult
	WorstPerformance  *PageResult
}

// ReportData represents the complete report structure
type ReportData struct {
	Generated time.Time     `json:"generated"`
	Summary   ReportSummary `json:"summary"`
	Results   []Result      `json:"results"`
}
