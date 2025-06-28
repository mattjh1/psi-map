package runner

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/logger"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils"
)

// RunBatch runs PSI tests concurrently for a list of URLs with limited concurrency.
func RunBatch(urls []string, maxConcurrent int) []*types.PageResult {
	// Get the singleton logger and configure it
	log := logger.GetLogger()

	// Create UI instance for rendering CLI elements
	ui := log.UI()

	var wg sync.WaitGroup
	results := make([]*types.PageResult, len(urls))
	sem := make(chan struct{}, maxConcurrent)
	var completed int32

	// Run spinner for progress feedback
	err := ui.RunProgressBar("Processing URLs", len(urls), func(increment func()) error {
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

				results[i] = &types.PageResult{
					URL:      url,
					Mobile:   &mobile,
					Desktop:  &desktop,
					Duration: duration,
				}

				// Log progress with Tagged logging
				log.Tagged("ANALYZE", "Processed URL: %s", "", url)
				increment()
				atomic.AddInt32(&completed, 1)
			}(i, url)
		}

		wg.Wait()
		return nil
	})
	if err != nil {
		log.Error("Failed to process URLs: %v", err)
		return results
	}

	// Log success
	log.Success("All tasks completed!")

	return results
}
