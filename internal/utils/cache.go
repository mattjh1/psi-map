package utils

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mattjh1/psi-map/internal/types"
)

// getCacheDir returns the appropriate cache directory for the OS
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
			return "", err
		}
		cacheDir = filepath.Join(homeDir, "Library", "Caches")
	default: // linux and others
		cacheDir = os.Getenv("XDG_CACHE_HOME")
		if cacheDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			cacheDir = filepath.Join(homeDir, ".cache")
		}
	}

	psiCacheDir := filepath.Join(cacheDir, "psi-map")

	// Ensure directory exists
	if err := os.MkdirAll(psiCacheDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %v", err)
	}

	return psiCacheDir, nil
}

// calculateSitemapHash calculates MD5 hash of sitemap content
func calculateSitemapHash(sitemapPath string) (string, error) {
	file, err := os.Open(sitemapPath)
	if err != nil {
		return "", fmt.Errorf("failed to open sitemap: %v", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %v", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// getCacheFilename returns the cache filename for a given sitemap hash
func getCacheFilename(cacheDir, hash string) string {
	return filepath.Join(cacheDir, fmt.Sprintf("sitemap-hash-%s.json", hash))
}

// CachedData represents cached results with metadata
type CachedData struct {
	Timestamp  time.Time          `json:"timestamp"`
	SitemapURL string             `json:"sitemap_url,omitempty"` // Optional: for human readability
	Results    []types.PageResult `json:"results"`
}

// CheckCache checks if cached results exist and are still valid for the given sitemap
func CheckCache(sitemapPath string, ttlHours int) ([]types.PageResult, bool, error) {
	// Calculate hash of current sitemap
	hash, err := calculateSitemapHash(sitemapPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to calculate sitemap hash: %v", err)
	}

	// Get cache directory
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, false, fmt.Errorf("failed to get cache directory: %v", err)
	}

	// Check if cache file exists
	cacheFile := getCacheFilename(cacheDir, hash)
	_, err = os.Stat(cacheFile)
	if os.IsNotExist(err) {
		return nil, false, nil // No cache found
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to stat cache file: %v", err)
	}

	// Load cached data
	cachedData, err := loadCachedData(cacheFile)
	if err != nil {
		return nil, false, fmt.Errorf("failed to load cached results: %v", err)
	}

	// Check TTL
	if ttlHours > 0 {
		expiryTime := cachedData.Timestamp.Add(time.Duration(ttlHours) * time.Hour)
		if time.Now().After(expiryTime) {
			// Cache expired, remove the file
			if err := os.Remove(cacheFile); err != nil {
				// Log but don't fail - we'll just regenerate
				fmt.Printf("Warning: failed to remove expired cache file: %v\n", err)
			}
			return nil, false, nil // Cache expired
		}
	}

	return cachedData.Results, true, nil
}

// SaveCache saves results to cache with hash-based filename and timestamp
func SaveCache(sitemapPath string, results []types.PageResult) error {
	// Calculate hash of sitemap
	hash, err := calculateSitemapHash(sitemapPath)
	if err != nil {
		return fmt.Errorf("failed to calculate sitemap hash: %v", err)
	}

	// Get cache directory
	cacheDir, err := getCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %v", err)
	}

	// Create cached data structure
	cachedData := CachedData{
		Timestamp:  time.Now(),
		SitemapURL: sitemapPath, // Store for human readability
		Results:    results,
	}

	// Create cache file
	cacheFile := getCacheFilename(cacheDir, hash)
	file, err := os.Create(cacheFile)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %v", err)
	}
	defer file.Close()

	// Encode results to JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cachedData); err != nil {
		return fmt.Errorf("failed to encode results: %v", err)
	}

	return nil
}

// loadCachedData loads CachedData from cache file with backward compatibility
func loadCachedData(filename string) (*CachedData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache file: %v", err)
	}
	defer file.Close()

	// Read the entire file to detect format
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %v", err)
	}

	// Try to unmarshal as new format first
	var cachedData CachedData
	if err := json.Unmarshal(data, &cachedData); err == nil {
		// Successfully parsed as new format
		return &cachedData, nil
	}

	// If that fails, try old format (direct array of PageResult)
	var results []types.PageResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to decode cache file in both old and new formats: %v", err)
	}

	// Convert old format to new format
	// Use file modification time as timestamp for old cache files
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	cachedData = CachedData{
		Timestamp:  fileInfo.ModTime(),
		SitemapURL: "unknown (legacy cache)", // We don't have this info in old format
		Results:    results,
	}

	return &cachedData, nil
}

// CacheInfo represents information about a cache file
type CacheInfo struct {
	Filename   string    `json:"filename"`
	Hash       string    `json:"hash"`
	SitemapURL string    `json:"sitemap_url"`
	Timestamp  time.Time `json:"timestamp"`
	Age        string    `json:"age"`
	IsExpired  bool      `json:"is_expired"`
}

// ListCacheFiles returns detailed information about cached files
func ListCacheFiles(ttlHours int) ([]CacheInfo, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %v", err)
	}

	var cacheInfos []CacheInfo
	now := time.Now()

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			cacheFile := filepath.Join(cacheDir, entry.Name())

			// Load cache data to get timestamp and sitemap info
			cachedData, err := loadCachedData(cacheFile)
			if err != nil {
				// Skip corrupted cache files
				continue
			}

			// Extract hash from filename (remove "sitemap-hash-" prefix and ".json" suffix)
			filename := entry.Name()
			hash := ""
			if len(filename) > 13 && filename[:13] == "sitemap-hash-" {
				hash = filename[13 : len(filename)-5] // Remove prefix and .json
			}

			age := now.Sub(cachedData.Timestamp)
			isExpired := ttlHours > 0 && age > time.Duration(ttlHours)*time.Hour

			cacheInfo := CacheInfo{
				Filename:   filename,
				Hash:       hash,
				SitemapURL: cachedData.SitemapURL,
				Timestamp:  cachedData.Timestamp,
				Age:        formatDuration(age),
				IsExpired:  isExpired,
			}

			cacheInfos = append(cacheInfos, cacheInfo)
		}
	}

	return cacheInfos, nil
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	} else {
		return fmt.Sprintf("%.1fd", d.Hours()/24)
	}
}

// CleanExpiredCache removes expired cache files
func CleanExpiredCache(ttlHours int) (int, error) {
	if ttlHours <= 0 {
		return 0, fmt.Errorf("TTL must be positive")
	}

	cacheInfos, err := ListCacheFiles(ttlHours)
	if err != nil {
		return 0, err
	}

	cacheDir, err := getCacheDir()
	if err != nil {
		return 0, err
	}

	cleaned := 0
	for _, info := range cacheInfos {
		if info.IsExpired {
			cacheFile := filepath.Join(cacheDir, info.Filename)
			if err := os.Remove(cacheFile); err != nil {
				return cleaned, fmt.Errorf("failed to remove expired cache file %s: %v", info.Filename, err)
			}
			cleaned++
		}
	}

	return cleaned, nil
}

// ClearCache removes all cache files
func ClearCache() error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			if err := os.Remove(filepath.Join(cacheDir, entry.Name())); err != nil {
				return fmt.Errorf("failed to remove cache file %s: %v", entry.Name(), err)
			}
		}
	}

	return nil
}
