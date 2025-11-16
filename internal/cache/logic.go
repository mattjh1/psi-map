package cache

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/mattjh1/psi-map/internal/types"
)

// LoadOrCreateIndex loads a SitemapIndex from the store using the given hash
// or creates a new one from orderedURLs and sitemapURL if not found.
func LoadOrCreateIndex(store CacheStore, hash string, orderedURLs []string, sitemapURL string) (*SitemapIndex, string, error) {
	filename := "sitemap-" + hash + ".json"
	idx, err := store.LoadSitemapIndexFile(filename)
	if err == nil {
		// ensure FileMap exists
		if idx.FileMap == nil {
			idx.FileMap = make(map[string]string, len(idx.URLs))
		}
		return idx, filename, nil
	}

	// Not found -> create
	idx = &SitemapIndex{
		Hash:       hash,
		SitemapURL: sitemapURL,
		URLs:       append([]string(nil), orderedURLs...), // copy
		FileMap:    make(map[string]string, len(orderedURLs)),
		UpdatedAt:  time.Now(),
	}
	if err := store.SaveSitemapIndexFile(filename, idx); err != nil {
		return nil, "", fmt.Errorf("failed to save new sitemap index: %w", err)
	}
	return idx, filename, nil
}

// CheckURLCache checks which URLs are present & fresh. It returns cached results (in sitemap order)
// and a slice of missing URLs (also in sitemap order). TTL is a duration (0 = no TTL).
func CheckURLCache(store CacheStore, index *SitemapIndex, ttl time.Duration) ([]*types.PageResult, []string, error) {
	now := time.Now()
	cached := make([]*types.PageResult, 0, len(index.URLs))
	missing := make([]string, 0, len(index.URLs))

	for _, url := range index.URLs {
		filename, ok := index.FileMap[url]
		if !ok {
			missing = append(missing, url)
			continue
		}
		// attempt to load entry
		entry, err := store.LoadURLFile(filename)
		if err != nil {
			// treat as missing if load fails
			missing = append(missing, url)
			continue
		}
		if ttl > 0 && now.Sub(entry.Timestamp) > ttl {
			// expired
			missing = append(missing, url)
			continue
		}
		cached = append(cached, entry.Result)
	}
	return cached, missing, nil
}

// SaveResults writes fresh results to the store and updates the index FileMap.
// It returns the updated index filename used when saving (so callers can persist index).
func SaveResults(store CacheStore, index *SitemapIndex, idxFilename string, fresh []*types.PageResult) (string, error) {
	for _, r := range fresh {
		// compute filename for URL
		filename := ComputeURLFilename(r.URL)
		entry := &CacheEntry{
			URL:       r.URL,
			Result:    r,
			Timestamp: time.Now(),
		}
		if err := store.SaveURLFile(filename, entry); err != nil {
			return "", fmt.Errorf("failed to save url file: %w", err)
		}
		index.FileMap[r.URL] = filepath.Base(filename)
	}
	index.UpdatedAt = time.Now()
	// persist index
	if err := store.SaveSitemapIndexFile(idxFilename, index); err != nil {
		return "", fmt.Errorf("failed to save sitemap index file: %w", err)
	}
	return idxFilename, nil
}

// CombineResultsInOrder merges cached and fresh maintaining the order from 'orderedURLs'.
// fresh takes precedence over cached for the same URL.
func CombineResultsInOrder(orderedURLs []string, cached, fresh []*types.PageResult) []*types.PageResult {
	lookup := make(map[string]*types.PageResult, len(cached)+len(fresh))
	for _, r := range cached {
		if r != nil {
			lookup[r.URL] = r
		}
	}
	for _, r := range fresh {
		if r != nil {
			lookup[r.URL] = r
		}
	}

	out := make([]*types.PageResult, 0, len(orderedURLs))
	for _, u := range orderedURLs {
		if r, ok := lookup[u]; ok {
			out = append(out, r)
		}
	}
	return out
}

// CleanExpired removes expired URL files and updates indexes. Returns number removed.
func CleanExpired(store CacheStore, ttl time.Duration, dryRun bool) (int, error) {
	if ttl <= 0 {
		return 0, fmt.Errorf("ttl must be positive")
	}
	removed := 0
	indexFiles, err := store.ListSitemapIndexFiles()
	if err != nil {
		return 0, fmt.Errorf("failed to list sitemap index files: %w", err)
	}

	now := time.Now()
	for _, idxFilename := range indexFiles {
		idx, err := store.LoadSitemapIndexFile(idxFilename)
		if err != nil {
			// skip bad index
			continue
		}

		updatedMap := make(map[string]string, len(idx.FileMap))
		for _, u := range idx.URLs {
			filename, ok := idx.FileMap[u]
			if !ok {
				continue
			}
			entry, err := store.LoadURLFile(filename)
			if err != nil {
				removed++
				continue
			}
			if now.Sub(entry.Timestamp) > ttl {
				removed++
				if !dryRun {
					_ = store.DeleteURLFile(filename)
				}
				continue
			}
			updatedMap[u] = filename
		}

		// write back index if changed and not dry run
		if !dryRun {
			if len(updatedMap) == 0 {
				// delete index file
				_ = store.DeleteSitemapIndexFile(idxFilename)
			} else {
				idx.FileMap = updatedMap
				idx.UpdatedAt = now
				_ = store.SaveSitemapIndexFile(idxFilename, idx)
			}
		}
	}
	return removed, nil
}
