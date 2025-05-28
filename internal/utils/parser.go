package utils

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mattjh1/psi-map/internal/types"
)

// ParseSitemap takes a path or URL to a sitemap and returns a slice of URLs
func ParseSitemap(input string) ([]string, error) {
	var reader io.Reader

	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		resp, err := http.Get(input)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("non-200 status: %d", resp.StatusCode)
		}
		reader = resp.Body
	} else {
		file, err := os.Open(input)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()
		reader = file
	}

	decoder := xml.NewDecoder(reader)
	var sitemap types.Sitemap
	if err := decoder.Decode(&sitemap); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	urls := make([]string, 0, len(sitemap.URLs))
	for _, u := range sitemap.URLs {
		urls = append(urls, strings.TrimSpace(u.Loc))
	}

	return urls, nil
}
