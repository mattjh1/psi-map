package cache

import (
	"time"

	"github.com/mattjh1/psi-map/internal/types"
)

// CacheEntry stores the cached result for a single URL.
type CacheEntry struct {
	URL       string            `json:"url"`
	Result    *types.PageResult `json:"result"`
	Timestamp time.Time         `json:"timestamp"`
}

// SitemapIndex stores the sitemap metadata and URL order.
type SitemapIndex struct {
	Hash       string            `json:"hash"`        // hash of sitemap (or generated hash from URL list)
	SitemapURL string            `json:"sitemap_url"` // original sitemap path/URL (may be empty)
	URLs       []string          `json:"urls"`        // **preserved sitemap order**
	FileMap    map[string]string `json:"file_map"`    // url -> filename (basename in urls dir)
	UpdatedAt  time.Time         `json:"updated_at"`
}
