package runner

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils"
)

// RunBatch runs PSI tests concurrently for a list of URLs with limited concurrency.
func RunBatch(urls []string, maxConcurrent int) []types.PageResult {
	var wg sync.WaitGroup
	results := make([]types.PageResult, len(urls))
	sem := make(chan struct{}, maxConcurrent)

	var completed int32 // for progress tracking

	for i, url := range urls {
		wg.Add(1)

		go func(i int, url string) {
			defer wg.Done()
			sem <- struct{}{}        // acquire slot
			defer func() { <-sem }() // release slot

			fmt.Printf("[Worker] Starting: %s\n", url)
			start := time.Now()

			var wgInner sync.WaitGroup
			var mobile, desktop types.Result

			wgInner.Add(2)

			go func() {
				defer wgInner.Done()
				fmt.Printf("  ↳ [Mobile] Fetching: %s\n", url)
				mobile = utils.FetchScore(url, "mobile")
			}()

			go func() {
				defer wgInner.Done()
				fmt.Printf("  ↳ [Desktop] Fetching: %s\n", url)
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

			fmt.Printf("[Worker] Finished: %s in %s\n", url, duration)
			atomic.AddInt32(&completed, 1)
		}(i, url)
	}

	// Progress monitor
	go func() {
		total := len(urls)
		for {
			time.Sleep(3 * time.Second)
			done := atomic.LoadInt32(&completed)
			fmt.Printf("[INFO] Progress: %d / %d completed\n", done, total)
			if int(done) == total {
				break
			}
		}
	}()

	wg.Wait()
	return results
}
