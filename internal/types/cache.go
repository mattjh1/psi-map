package types

import "time"

// CachedData represents cached results with metadata
type CachedData struct {
	Timestamp  time.Time    `json:"timestamp"`
	SitemapURL string       `json:"sitemap_url,omitempty"` // Optional: for human readability
	Results    []PageResult `json:"results"`
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
