package utils

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils/validate"
)

func fetchRemoteSitemap(input string) (io.ReadCloser, error) {
	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("non-200 status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// ParseSitemap takes a path or URL to a sitemap and returns a slice of URLs
func ParseSitemap(input string) ([]string, error) {
	var reader io.ReadCloser

	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		body, err := fetchRemoteSitemap(input)
		if err != nil {
			return nil, err
		}
		defer body.Close()
		reader = body
	} else {
		file, err := validate.SafeOpenFile(input)
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
