package server

import (
	"fmt"
	"time"

	"github.com/mattjh1/psi-map/internal/types"
)

// Test data helpers
func createMockResult(url string, perfScore, accScore, bpScore, seoScore float64, hasError bool) types.PageResult {
	var mobileErr, desktopErr error
	if hasError {
		mobileErr = fmt.Errorf("mock error")
		desktopErr = fmt.Errorf("mock error")
	}

	scores := &types.CategoryScores{
		Performance:   perfScore,
		Accessibility: accScore,
		BestPractices: bpScore,
		SEO:           seoScore,
	}

	return types.PageResult{
		URL: url,
		Mobile: types.Result{
			Scores: scores,
			Error:  mobileErr,
		},
		Desktop: types.Result{
			Scores: scores,
			Error:  desktopErr,
		},
		Duration: time.Duration(100) * time.Millisecond,
	}
}

func createResultWithMetrics(url string, perfScore float64) types.PageResult {
	return types.PageResult{
		URL: url,
		Mobile: types.Result{
			Scores: &types.CategoryScores{
				Performance:   perfScore,
				Accessibility: 90,
				BestPractices: 85,
				SEO:           95,
			},
			Metrics: &types.Metrics{
				FirstContentfulPaint:   1200,
				LargestContentfulPaint: 2500,
				CumulativeLayoutShift:  0.1,
				FirstInputDelay:        50,
				SpeedIndex:             2800,
				TimeToInteractive:      3200,
				TotalBlockingTime:      150,
				DOMSize:                1200,
				ResourceCount:          45,
				TransferSize:           512000,
			},
			Error: nil,
		},
		Desktop: types.Result{
			Scores: &types.CategoryScores{
				Performance:   perfScore,
				Accessibility: 95,
				BestPractices: 90,
				SEO:           98,
			},
			Error: nil,
		},
		Duration: time.Duration(150) * time.Millisecond,
	}
}

func createResultWithPartialSuccess(url string, mobileSuccess, desktopSuccess bool) types.PageResult {
	var mobileErr, desktopErr error
	var mobileScores, desktopScores *types.CategoryScores

	if mobileSuccess {
		mobileScores = &types.CategoryScores{
			Performance:   80,
			Accessibility: 85,
			BestPractices: 90,
			SEO:           95,
		}
	} else {
		mobileErr = fmt.Errorf("mobile error")
	}

	if desktopSuccess {
		desktopScores = &types.CategoryScores{
			Performance:   85,
			Accessibility: 90,
			BestPractices: 95,
			SEO:           98,
		}
	} else {
		desktopErr = fmt.Errorf("desktop error")
	}

	return types.PageResult{
		URL: url,
		Mobile: types.Result{
			Scores: mobileScores,
			Error:  mobileErr,
		},
		Desktop: types.Result{
			Scores: desktopScores,
			Error:  desktopErr,
		},
		Duration: time.Duration(200) * time.Millisecond,
	}
}
