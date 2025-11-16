package cache

// CacheStore is the storage backend abstraction.
// Implementations must be safe for concurrent use if used concurrently.
type CacheStore interface {
	// URL entries
	SaveURLFile(filename string, entry *CacheEntry) error
	LoadURLFile(filename string) (*CacheEntry, error)
	DeleteURLFile(filename string) error

	// Sitemap index
	SaveSitemapIndexFile(filename string, index *SitemapIndex) error
	LoadSitemapIndexFile(filename string) (*SitemapIndex, error)
	DeleteSitemapIndexFile(filename string) error

	// Helpers
	ListSitemapIndexFiles() ([]string, error)
	Close() error
}
