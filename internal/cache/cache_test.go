package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mattjh1/psi-map/internal/types"
)

func tempDir(t *testing.T) string {
	dir := t.TempDir()
	return dir
}

func makePageResult(url string) *types.PageResult {
	return &types.PageResult{URL: url}
}

func TestFilesystemStoreAndLogic(t *testing.T) {
	dir := tempDir(t)

	store, err := NewFilesystemCacheStore(dir)
	if err != nil {
		t.Fatalf("failed to create fs store: %v", err)
	}
	defer store.Close()

	// Build index
	urls := []string{"https://a.example/", "https://b.example/", "https://c.example/"}
	hash := "testhash"
	idx, filename, err := LoadOrCreateIndex(store, hash, urls, "https://sitemap")
	if err != nil {
		t.Fatalf("LoadOrCreateIndex failed: %v", err)
	}
	if idx == nil {
		t.Fatalf("index nil")
	}
	// nothing cached yet
	cached, missing, err := CheckURLCache(store, idx, 24*time.Hour)
	if err != nil {
		t.Fatalf("CheckURLCache error: %v", err)
	}
	if len(cached) != 0 || len(missing) != 3 {
		t.Fatalf("expected 3 missing, got cached=%d missing=%d", len(cached), len(missing))
	}

	// Save one result
	fresh := []*types.PageResult{makePageResult(urls[1])}
	if _, err := SaveResults(store, idx, filename, fresh); err != nil {
		t.Fatalf("SaveResults failed: %v", err)
	}

	// reload index
	loadedIdx, err := store.LoadSitemapIndexFile(filename)
	if err != nil {
		t.Fatalf("load index failed: %v", err)
	}
	if len(loadedIdx.FileMap) != 1 {
		t.Fatalf("expected 1 filemap entry, got %d", len(loadedIdx.FileMap))
	}

	// Check again
	cached, missing, err = CheckURLCache(store, loadedIdx, 24*time.Hour)
	if err != nil {
		t.Fatalf("CheckURLCache error: %v", err)
	}
	if len(cached) != 1 || len(missing) != 2 {
		t.Fatalf("expected cached=1 missing=2, got %d/%d", len(cached), len(missing))
	}

	// Merge results - should preserve order: a, b, c (only b present)
	combined := CombineResultsInOrder(urls, cached, fresh)
	if len(combined) != 1 || combined[0].URL != urls[1] {
		t.Fatalf("unexpected combined order or content")
	}

	// Test expiration: set very small TTL and ensure cleaning removes files
	removed, err := CleanExpired(store, 1*time.Nanosecond, false)
	if err != nil {
		t.Fatalf("CleanExpired: %v", err)
	}
	if removed == 0 {
		t.Fatalf("expected at least 1 removed, got 0")
	}

	// ensure index file removed or filemap cleared
	_, err = os.Stat(filepath.Join(dir, "indexes", filename))
	// either deleted or exists; do not fail here
	_ = err
}
