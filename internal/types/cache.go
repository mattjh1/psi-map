// Add these new types to your types.go file
package types

import (
	"time"
)

// URLCacheDetail represents detailed information about a cached URL
type URLCacheDetail struct {
	URL              string    `json:"url"`
	Age              string    `json:"age"`
	IsExpired        bool      `json:"is_expired"`
	IsStale          bool      `json:"is_stale"` // > 50% of TTL
	PerformanceScore float64   `json:"performance_score,omitempty"`
	CacheSize        int64     `json:"cache_size"`
	Timestamp        time.Time `json:"timestamp"`
	HasErrors        bool      `json:"has_errors"`
}

// Add these fields to your existing CacheInfo struct
type CacheInfo struct {
	Filename     string    `json:"filename"`
	Hash         string    `json:"hash"`
	FullHash     string    `json:"full_hash"`
	SitemapURL   string    `json:"sitemap_url"`
	Timestamp    time.Time `json:"timestamp"`
	Age          string    `json:"age"`
	IsExpired    bool      `json:"is_expired"`
	URLCount     int       `json:"url_count"`
	ValidCount   int       `json:"valid_count"`
	ExpiredCount int       `json:"expired_count"`
	// New fields for enhanced details
	StaleCount int     `json:"stale_count,omitempty"`
	TotalSize  int64   `json:"total_size,omitempty"`
	AvgScore   float64 `json:"avg_score,omitempty"`
}

// Add these functions to your utils.go file

// GetURLCacheDetails returns detailed information about URLs in a specific sitemap cache
