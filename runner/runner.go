package runner

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils"
	"github.com/pterm/pterm"
)

// RunBatch runs PSI tests concurrently for a list of URLs with limited concurrency.
func RunBatch(urls []string, maxConcurrent int) []types.PageResult {
	var wg sync.WaitGroup
	results := make([]types.PageResult, len(urls))
	sem := make(chan struct{}, maxConcurrent)
	var completed int32

	// Create progress bar
	var progressbar *pterm.ProgressbarPrinter
	progressbar, _ = pterm.DefaultProgressbar.WithTotal(len(urls)).WithTitle("Processing URLs").Start()

	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			start := time.Now()
			var wgInner sync.WaitGroup
			var mobile, desktop types.Result

			wgInner.Add(constants.WaitGroupWorkers)
			go func() {
				defer wgInner.Done()
				mobile = utils.FetchScore(url, "mobile")
			}()

			go func() {
				defer wgInner.Done()
				desktop = utils.FetchScore(url, "desktop")
			}()

			wgInner.Wait()
			duration := time.Since(start)

			results[i] = types.PageResult{
				URL:      url,
				Mobile:   mobile,
				Desktop:  desktop,
				Duration: duration,
			}

			// Update progress
			atomic.AddInt32(&completed, 1)
			progressbar.UpdateTitle("Processing URLs - " + url)
			progressbar.Increment()
		}(i, url)
	}

	wg.Wait()

	pterm.Success.Printf("All tasks completed!")

	return results
}
