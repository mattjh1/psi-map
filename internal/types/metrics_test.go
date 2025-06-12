package types

import (
	"testing"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/stretchr/testify/assert"
)

// TestGetCoreWebVitalsGrade tests the GetCoreWebVitalsGrade method for various metric values
func TestGetCoreWebVitalsGrade(t *testing.T) {
	tests := []struct {
		name     string
		metrics  Metrics
		expected map[string]string
	}{
		{
			name: "All metrics in Good range",
			metrics: Metrics{
				FirstContentfulPaint:   constants.FCPGoodThreshold - 100,
				LargestContentfulPaint: constants.LCPGoodThreshold - 100,
				CumulativeLayoutShift:  constants.CLSGoodThreshold - 0.01,
				FirstInputDelay:        constants.FIDGoodThreshold - 10,
			},
			expected: map[string]string{
				"fcp": constants.GradeGood,
				"lcp": constants.GradeGood,
				"cls": constants.GradeGood,
				"fid": constants.GradeGood,
			},
		},
		{
			name: "All metrics in Needs Improvement range",
			metrics: Metrics{
				FirstContentfulPaint:   constants.FCPGoodThreshold + 100,
				LargestContentfulPaint: constants.LCPGoodThreshold + 100,
				CumulativeLayoutShift:  constants.CLSGoodThreshold + 0.05,
				FirstInputDelay:        constants.FIDGoodThreshold + 10,
			},
			expected: map[string]string{
				"fcp": constants.GradeNeedsImprovement,
				"lcp": constants.GradeNeedsImprovement,
				"cls": constants.GradeNeedsImprovement,
				"fid": constants.GradeNeedsImprovement,
			},
		},
		{
			name: "All metrics in Poor range",
			metrics: Metrics{
				FirstContentfulPaint:   constants.FCPPoorThreshold + 100,
				LargestContentfulPaint: constants.LCPPoorThreshold + 100,
				CumulativeLayoutShift:  constants.CLSPoorThreshold + 0.05,
				FirstInputDelay:        constants.FIDPoorThreshold + 10,
			},
			expected: map[string]string{
				"fcp": constants.GradePoor,
				"lcp": constants.GradePoor,
				"cls": constants.GradePoor,
				"fid": constants.GradePoor,
			},
		},
		{
			name: "Mixed range metrics",
			metrics: Metrics{
				FirstContentfulPaint:   constants.FCPGoodThreshold - 100,
				LargestContentfulPaint: constants.LCPGoodThreshold + 100,
				CumulativeLayoutShift:  constants.CLSPoorThreshold + 0.05,
				FirstInputDelay:        constants.FIDGoodThreshold - 10,
			},
			expected: map[string]string{
				"fcp": constants.GradeGood,
				"lcp": constants.GradeNeedsImprovement,
				"cls": constants.GradePoor,
				"fid": constants.GradeGood,
			},
		},
		{
			name: "Boundary values for Good thresholds",
			metrics: Metrics{
				FirstContentfulPaint:   constants.FCPGoodThreshold,
				LargestContentfulPaint: constants.LCPGoodThreshold,
				CumulativeLayoutShift:  constants.CLSGoodThreshold,
				FirstInputDelay:        constants.FIDGoodThreshold,
			},
			expected: map[string]string{
				"fcp": constants.GradeNeedsImprovement,
				"lcp": constants.GradeNeedsImprovement,
				"cls": constants.GradeNeedsImprovement,
				"fid": constants.GradeNeedsImprovement,
			},
		},
		{
			name: "Boundary values for Poor thresholds",
			metrics: Metrics{
				FirstContentfulPaint:   constants.FCPPoorThreshold,
				LargestContentfulPaint: constants.LCPPoorThreshold,
				CumulativeLayoutShift:  constants.CLSPoorThreshold,
				FirstInputDelay:        constants.FIDPoorThreshold,
			},
			expected: map[string]string{
				"fcp": constants.GradePoor,
				"lcp": constants.GradePoor,
				"cls": constants.GradePoor,
				"fid": constants.GradePoor,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metrics.GetCoreWebVitalsGrade()
			assert.Equal(t, tt.expected, got, "Unexpected grades for metrics")
		})
	}
}
