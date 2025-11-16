package cache

import (
	"crypto/md5" // #nosec G401, G501 - used for hashing paths only
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattjh1/psi-map/internal/constants"
)

type FilesystemCacheStore struct {
	baseDir string
	urlsDir string
	idxDir  string
}

func NewFilesystemCacheStore(baseDir string) (*FilesystemCacheStore, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("baseDir required")
	}
	urlsDir := filepath.Join(baseDir, "urls")
	idxDir := filepath.Join(baseDir, "indexes")

	if err := os.MkdirAll(urlsDir, constants.DefaultDirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create urls dir: %w", err)
	}
	if err := os.MkdirAll(idxDir, constants.DefaultDirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create indexes dir: %w", err)
	}

	return &FilesystemCacheStore{baseDir: baseDir, urlsDir: urlsDir, idxDir: idxDir}, nil
}

func (fs *FilesystemCacheStore) Close() error { return nil }

func filenameForURL(url string) string {
	sum := md5.Sum([]byte(url)) // #nosec G401 - non-crypto hash
	return "url-" + hex.EncodeToString(sum[:]) + ".json"
}

func (fs *FilesystemCacheStore) urlFilePath(filename string) string {
	return filepath.Join(fs.urlsDir, filename)
}

func (fs *FilesystemCacheStore) idxFilePath(filename string) string {
	return filepath.Join(fs.idxDir, filename)
}

// SaveURLFile saves a CacheEntry to the given filename (basename).
func (fs *FilesystemCacheStore) SaveURLFile(filename string, entry *CacheEntry) error {
	if filename == "" {
		filename = filenameForURL(entry.URL)
	}
	full := fs.urlFilePath(filename)
	tmp := full + ".tmp"

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	if err := os.WriteFile(tmp, data, constants.DefaultFilePermissions); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	if err := os.Rename(tmp, full); err != nil {
		return fmt.Errorf("failed to rename cache file: %w", err)
	}
	return nil
}

func (fs *FilesystemCacheStore) LoadURLFile(filename string) (*CacheEntry, error) {
	full := fs.urlFilePath(filename)
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}
	var e CacheEntry
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}
	return &e, nil
}

func (fs *FilesystemCacheStore) DeleteURLFile(filename string) error {
	full := fs.urlFilePath(filename)
	_ = os.Remove(full)
	return nil
}

func (fs *FilesystemCacheStore) SaveSitemapIndexFile(filename string, index *SitemapIndex) error {
	if filename == "" {
		filename = "sitemap-" + index.Hash + ".json"
	}
	full := fs.idxFilePath(filename)
	tmp := full + ".tmp"

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sitemap index: %w", err)
	}

	if err := os.WriteFile(tmp, data, constants.DefaultFilePermissions); err != nil {
		return fmt.Errorf("failed to write sitemap index file: %w", err)
	}
	if err := os.Rename(tmp, full); err != nil {
		return fmt.Errorf("failed to rename sitemap index file: %w", err)
	}
	return nil
}

func (fs *FilesystemCacheStore) LoadSitemapIndexFile(filename string) (*SitemapIndex, error) {
	full := fs.idxFilePath(filename)
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitemap index file: %w", err)
	}
	var idx SitemapIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sitemap index: %w", err)
	}
	return &idx, nil
}

func (fs *FilesystemCacheStore) DeleteSitemapIndexFile(filename string) error {
	full := fs.idxFilePath(filename)
	_ = os.Remove(full)
	return nil
}

func (fs *FilesystemCacheStore) ListSitemapIndexFiles() ([]string, error) {
	entries, err := os.ReadDir(fs.idxDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitemap index dir: %w", err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		out = append(out, e.Name())
	}
	return out, nil
}

// Public helper to compute url filename (string)
func ComputeURLFilename(urlStr string) string {
	return filenameForURL(urlStr)
}
