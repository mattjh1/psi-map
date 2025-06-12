package types

import "github.com/mattjh1/psi-map/internal/constants"

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
	case m.FirstContentfulPaint < constants.FCPGoodThreshold:
		grades["fcp"] = constants.GradeGood
	case m.FirstContentfulPaint < constants.FCPPoorThreshold:
		grades["fcp"] = constants.GradeNeedsImprovement
	default:
		grades["fcp"] = constants.GradePoor
	}

	// LCP grading (< 2.5s = good, < 4s = needs improvement, >= 4s = poor)
	switch {
	case m.LargestContentfulPaint < constants.LCPGoodThreshold:
		grades["lcp"] = constants.GradeGood
	case m.LargestContentfulPaint < constants.LCPPoorThreshold:
		grades["lcp"] = constants.GradeNeedsImprovement
	default:
		grades["lcp"] = constants.GradePoor
	}

	// CLS grading (< 0.1 = good, < 0.25 = needs improvement, >= 0.25 = poor)
	switch {
	case m.CumulativeLayoutShift < constants.CLSGoodThreshold:
		grades["cls"] = constants.GradeGood
	case m.CumulativeLayoutShift < constants.CLSPoorThreshold:
		grades["cls"] = constants.GradeNeedsImprovement
	default:
		grades["cls"] = constants.GradePoor
	}

	// FID grading (< 100ms = good, < 300ms = needs improvement, >= 300ms = poor)
	switch {
	case m.FirstInputDelay < constants.FIDGoodThreshold:
		grades["fid"] = constants.GradeGood
	case m.FirstInputDelay < constants.FIDPoorThreshold:
		grades["fid"] = constants.GradeNeedsImprovement
	default:
		grades["fid"] = constants.GradePoor
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
