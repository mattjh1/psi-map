package utils

import (
	// #nosec G501 - used only for checksums, not for security
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/utils/validate"
)

func GetCacheDir() (string, error) {
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

func CalculateSitemapHash(sitemapPath string, urls []string) (string, error) {
	// #nosec G401 - used only for checksums, not for security
	hash := md5.New()
	if sitemapPath != "" && (strings.HasPrefix(sitemapPath, "http://") || strings.HasPrefix(sitemapPath, "https://")) {
		// For remote sitemaps, hash the URL itself
		hash.Write([]byte(sitemapPath))
	} else if sitemapPath != "" {
		// For local files, hash the content
		file, err := validate.SafeOpenFile(sitemapPath)
		if err != nil {
			// If file doesn't exist, it might be a list of URLs. Hash the path itself as a fallback.
			hash.Write([]byte(sitemapPath))
		} else {
			defer file.Close()
			if _, err := io.Copy(hash, file); err != nil {
				return "", fmt.Errorf("failed to calculate hash: %w", err)
			}
		}
	} else {
		// For URL lists provided directly, hash the URLs
		for _, url := range urls {
			hash.Write([]byte(url + "\n"))
		}
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
