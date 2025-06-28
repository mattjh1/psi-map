package utils

import (
	// #nosec G501 - used only for checksums, not for security
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/types"
)

// URLCacheEntry represents a cached result for a single URL
type URLCacheEntry struct {
	URL        string           `json:"url"`
	Result     types.PageResult `json:"result"`
	Timestamp  time.Time        `json:"timestamp"`
	SitemapURL string           `json:"sitemap_url"`
}

// SitemapCacheIndex tracks which URLs belong to which sitemap
type SitemapCacheIndex struct {
	SitemapURL  string            `json:"sitemap_url"`
	SitemapHash string            `json:"sitemap_hash"`
	URLs        map[string]string `json:"urls"`
	LastUpdated time.Time         `json:"last_updated"`
}

var CacheDir = getCacheDir

func getCacheDir() (string, error) {
	var cacheDir string
	switch runtime.GOOS {
	case "windows":
		cacheDir = os.Getenv("LOCALAPPDATA")
		if cacheDir == "" {
			cacheDir = os.Getenv("TEMP")
		}
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		cacheDir = filepath.Join(homeDir, "Library", "Caches")
	default:
		cacheDir = os.Getenv("XDG_CACHE_HOME")
		if cacheDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %v", err)
			}
			cacheDir = filepath.Join(homeDir, ".cache")
		}
	}

	psiCacheDir := filepath.Join(cacheDir, "psi-map")
	if err := os.MkdirAll(psiCacheDir, constants.DefaultDirPermissions); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %v", err)
	}
	return psiCacheDir, nil
}

func calculateSitemapHash(sitemapPath string, urls []string) (string, error) {
	// #nosec G401 - used only for checksums, not for security
	hash := md5.New()
	if sitemapPath != "" {
		file, err := os.Open(sitemapPath)
		if err != nil {
			return "", fmt.Errorf("failed to open sitemap: %v", err)
		}
		defer file.Close()
		if _, err := io.Copy(hash, file); err != nil {
			return "", fmt.Errorf("failed to calculate hash: %v", err)
		}
	} else {
		for _, url := range urls {
			hash.Write([]byte(url + "\n"))
		}
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func getURLCacheFilename(cacheDir, url string) string {
	// #nosec G401 - used only for checksums, not for security
	urlHash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	return filepath.Join(cacheDir, "urls", fmt.Sprintf("url-%s.json", urlHash))
}

func getSitemapIndexFilename(cacheDir, sitemapHash string) string {
	return filepath.Join(cacheDir, "indexes", fmt.Sprintf("sitemap-%s.json", sitemapHash))
}

func loadSitemapIndex(filename string) (*SitemapCacheIndex, bool) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, false
	}
	defer file.Close()

	var index SitemapCacheIndex
	if err := json.NewDecoder(file).Decode(&index); err != nil {
		return nil, false
	}
	return &index, true
}

func saveSitemapIndex(filename string, index *SitemapCacheIndex) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create index file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(index); err != nil {
		return fmt.Errorf("failed to encode sitemap index to %s: %w", filename, err)
	}
	return nil
}

func loadURLCacheEntry(filename string) (*URLCacheEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load file from filesystem %s: %w", filename, err)
	}
	defer file.Close()

	var entry URLCacheEntry
	if err := json.NewDecoder(file).Decode(&entry); err != nil {
		return nil, fmt.Errorf("failed to load URL cache entry %s: %w", filename, err)
	}
	return &entry, nil
}

func saveURLCacheEntry(filename string, entry *URLCacheEntry) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to load file from filesystem %s: %w", filename, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(entry); err != nil {
		return fmt.Errorf("failed to save URL to cache %s: %w", filename, err)
	}
	return nil
}

func CheckURLCache(sitemapPath string, urls []string, ttlHours int) ([]*types.PageResult, []string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, urls, err
	}

	urlsCacheDir := filepath.Join(cacheDir, "urls")
	indexesCacheDir := filepath.Join(cacheDir, "indexes")
	if err := os.MkdirAll(urlsCacheDir, constants.DefaultDirPermissions); err != nil {
		return nil, urls, fmt.Errorf("failed to create URLs cache directory: %v", err)
	}
	if err := os.MkdirAll(indexesCacheDir, constants.DefaultDirPermissions); err != nil {
		return nil, urls, fmt.Errorf("failed to create indexes cache directory: %v", err)
	}

	currentHash, err := calculateSitemapHash(sitemapPath, urls)
	if err != nil {
		return nil, urls, err
	}

	indexFile := getSitemapIndexFilename(cacheDir, currentHash)
	index, indexExists := loadSitemapIndex(indexFile)
	if !indexExists {
		return nil, urls, nil
	}

	now := time.Now()
	cached := make([]*types.PageResult, 0)
	missing := make([]string, 0)

	for _, url := range urls {
		cacheFilename, exists := index.URLs[url]
		if !exists {
			missing = append(missing, url)
			continue
		}

		cacheFile := filepath.Join(cacheDir, "urls", cacheFilename)
		entry, err := loadURLCacheEntry(cacheFile)
		if err != nil {
			missing = append(missing, url)
			continue
		}

		if ttlHours > 0 {
			expiryTime := entry.Timestamp.Add(time.Duration(ttlHours) * time.Hour)
			if now.After(expiryTime) {
				missing = append(missing, url)
				os.Remove(cacheFile)
				continue
			}
		}
		cached = append(cached, &entry.Result)
	}

	return cached, missing, nil
}

func SaveURLCache(sitemapPath string, allURLs []string, newResults []*types.PageResult) error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	sitemapHash, err := calculateSitemapHash(sitemapPath, allURLs)
	if err != nil {
		return err
	}

	indexFile := getSitemapIndexFilename(cacheDir, sitemapHash)
	index, _ := loadSitemapIndex(indexFile)
	if index == nil {
		index = &SitemapCacheIndex{
			SitemapURL:  sitemapPath,
			SitemapHash: sitemapHash,
			URLs:        make(map[string]string),
			LastUpdated: time.Now(),
		}
	}

	for _, result := range newResults {
		entry := URLCacheEntry{
			URL:        result.URL,
			Result:     *result,
			Timestamp:  time.Now(),
			SitemapURL: sitemapPath,
		}

		cacheFile := getURLCacheFilename(cacheDir, result.URL)
		filename := filepath.Base(cacheFile)

		if err := saveURLCacheEntry(cacheFile, &entry); err != nil {
			return fmt.Errorf("failed to save cache entry for %s: %v", result.URL, err)
		}
		index.URLs[result.URL] = filename
	}

	index.LastUpdated = time.Now()
	return saveSitemapIndex(indexFile, index)
}

func ListCacheFiles(ttlHours int, verbose bool) ([]types.CacheInfo, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	indexesDir := filepath.Join(cacheDir, "indexes")
	entries, err := os.ReadDir(indexesDir)
	if err != nil {
		return []types.CacheInfo{}, nil
	}

	cacheInfos := make([]types.CacheInfo, 0, len(entries))
	now := time.Now()
	stalePeriod := time.Duration(float64(ttlHours)*0.5) * time.Hour

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "sitemap-") {
			continue
		}

		indexFile := filepath.Join(indexesDir, entry.Name())
		index, exists := loadSitemapIndex(indexFile)
		if !exists {
			continue
		}

		cacheInfo := types.CacheInfo{
			Filename:   entry.Name(),
			Hash:       index.SitemapHash[:8],
			FullHash:   index.SitemapHash,
			SitemapURL: index.SitemapURL,
			Timestamp:  index.LastUpdated,
			URLCount:   len(index.URLs),
		}

		if verbose {
			validCount, expiredCount, staleCount, totalSize, totalScore, avgScore := 0, 0, 0, int64(0), 0.0, 0.0
			scoreCount := 0

			for url := range index.URLs {
				cacheFile := filepath.Join(cacheDir, "urls", index.URLs[url])
				urlEntry, err := loadURLCacheEntry(cacheFile)
				if err != nil {
					expiredCount++
					continue
				}

				if fileInfo, err := os.Stat(cacheFile); err == nil {
					totalSize += fileInfo.Size()
				}

				if score := extractPerformanceScore(&urlEntry.Result); score > 0 {
					totalScore += score
					scoreCount++
				}

				if ttlHours > 0 {
					expiryTime := urlEntry.Timestamp.Add(time.Duration(ttlHours) * time.Hour)
					if now.After(expiryTime) {
						expiredCount++
					} else if now.After(urlEntry.Timestamp.Add(stalePeriod)) {
						staleCount++
					} else {
						validCount++
					}
				} else {
					validCount++
				}
			}

			if scoreCount > 0 {
				avgScore = totalScore / float64(scoreCount)
			}

			cacheInfo.ValidCount = validCount
			cacheInfo.ExpiredCount = expiredCount
			cacheInfo.StaleCount = staleCount
			cacheInfo.TotalSize = totalSize
			cacheInfo.AvgScore = avgScore
			cacheInfo.IsExpired = expiredCount > 0
		}

		cacheInfo.Age = formatDuration(now.Sub(index.LastUpdated))
		cacheInfos = append(cacheInfos, cacheInfo)
	}

	return cacheInfos, nil
}

func CleanExpiredCacheFiles(ttlHours int, dryRun bool) (int, error) {
	if ttlHours <= 0 {
		return 0, fmt.Errorf("TTL must be positive")
	}

	cacheDir, err := getCacheDir()
	if err != nil {
		return 0, err
	}

	indexesDir := filepath.Join(cacheDir, "indexes")
	entries, err := os.ReadDir(indexesDir)
	if err != nil {
		return 0, nil
	}

	cleaned := 0
	now := time.Now()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "sitemap-") {
			continue
		}

		indexFile := filepath.Join(indexesDir, entry.Name())
		index, exists := loadSitemapIndex(indexFile)
		if !exists {
			continue
		}

		updatedURLs := make(map[string]string)
		urlsRemoved := 0

		for url, filename := range index.URLs {
			cacheFile := filepath.Join(cacheDir, "urls", filename)
			urlEntry, err := loadURLCacheEntry(cacheFile)
			if err != nil {
				urlsRemoved++
				continue
			}

			expiryTime := urlEntry.Timestamp.Add(time.Duration(ttlHours) * time.Hour)
			if now.After(expiryTime) {
				if !dryRun {
					if err := os.Remove(cacheFile); err == nil {
						urlsRemoved++
					}
				} else {
					urlsRemoved++
				}
			} else {
				updatedURLs[url] = filename
			}
		}

		cleaned += urlsRemoved
		if !dryRun && len(updatedURLs) != len(index.URLs) {
			if len(updatedURLs) == 0 {
				os.Remove(indexFile)
			} else {
				index.URLs = updatedURLs
				index.LastUpdated = now
				if err := saveSitemapIndex(indexFile, index); err != nil {
					return -1, fmt.Errorf("warning: failed to save sitemap index: %w", err)
				}
			}
		}
	}

	return cleaned, nil
}

func ClearAllCacheFiles() (int, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return 0, err
	}

	clearedCount := 0
	urlsDir := filepath.Join(cacheDir, "urls")
	indexesDir := filepath.Join(cacheDir, "indexes")

	if urlEntries, err := os.ReadDir(urlsDir); err == nil {
		for _, entry := range urlEntries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "url-") && strings.HasSuffix(entry.Name(), ".json") {
				if err := os.Remove(filepath.Join(urlsDir, entry.Name())); err == nil {
					clearedCount++
				}
			}
		}
		os.RemoveAll(urlsDir)
	}

	if indexEntries, err := os.ReadDir(indexesDir); err == nil {
		for _, entry := range indexEntries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "sitemap-") && strings.HasSuffix(entry.Name(), ".json") {
				if err := os.Remove(filepath.Join(indexesDir, entry.Name())); err == nil {
					clearedCount++
				}
			}
		}
		os.RemoveAll(indexesDir)
	}

	return clearedCount, nil
}

func GetURLCacheDetails(sitemapHash string, ttlHours int) ([]types.URLCacheDetail, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	indexFile := getSitemapIndexFilename(cacheDir, sitemapHash)
	index, exists := loadSitemapIndex(indexFile)
	if !exists {
		return nil, fmt.Errorf("sitemap index not found")
	}

	details := make([]types.URLCacheDetail, 0, len(index.URLs))
	now := time.Now()
	stalePeriod := time.Duration(float64(ttlHours)*0.5) * time.Hour

	for url, filename := range index.URLs {
		cacheFile := filepath.Join(cacheDir, "urls", filename)
		urlEntry, err := loadURLCacheEntry(cacheFile)
		if err != nil {
			continue
		}

		fileInfo, err := os.Stat(cacheFile)
		cacheSize := int64(0)
		if err == nil {
			cacheSize = fileInfo.Size()
		}

		age := now.Sub(urlEntry.Timestamp)
		isExpired := false
		isStale := false

		if ttlHours > 0 {
			expiryTime := urlEntry.Timestamp.Add(time.Duration(ttlHours) * time.Hour)
			isExpired = now.After(expiryTime)
			if !isExpired {
				isStale = now.After(urlEntry.Timestamp.Add(stalePeriod))
			}
		}

		detail := types.URLCacheDetail{
			URL:              url,
			Age:              formatDuration(age),
			IsExpired:        isExpired,
			IsStale:          isStale,
			PerformanceScore: extractPerformanceScore(&urlEntry.Result),
			CacheSize:        cacheSize,
			Timestamp:        urlEntry.Timestamp,
			HasErrors:        hasErrors(&urlEntry.Result),
		}
		details = append(details, detail)
	}

	sort.Slice(details, func(i, j int) bool {
		return details[i].Timestamp.After(details[j].Timestamp)
	})

	return details, nil
}

func extractPerformanceScore(result *types.PageResult) float64 {
	if result == nil {
		return 0.0
	}
	if result.Mobile != nil && result.Mobile.Scores != nil && result.Mobile.Scores.Performance > 0 {
		return result.Mobile.Scores.Performance
	}
	if result.Desktop != nil && result.Desktop.Scores != nil && result.Desktop.Scores.Performance > 0 {
		return result.Desktop.Scores.Performance
	}
	return 0.0
}

func hasErrors(result *types.PageResult) bool {
	if result == nil {
		return true // Consider a nil result as having errors
	}
	if result.Mobile != nil && result.Mobile.Error != nil || result.Desktop != nil && result.Desktop.Error != nil {
		return true
	}
	if result.Mobile == nil && result.Desktop == nil {
		return true
	}
	if result.Mobile != nil && result.Mobile.Scores != nil && result.Mobile.Scores.Performance < 50 {
		return true
	}
	if result.Desktop != nil && result.Desktop.Scores != nil && result.Desktop.Scores.Performance < 50 {
		return true
	}
	return false
}

func formatDuration(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%.0fm", d.Minutes())
	case d < 24*time.Hour:
		return fmt.Sprintf("%.1fh", d.Hours())
	default:
		return fmt.Sprintf("%.1fd", d.Hours()/constants.Day24H)
	}
}
