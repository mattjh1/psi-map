package types

import (
	"time"
)

// Sitemap represents the structure of a basic XML sitemap
type Sitemap struct {
	URLs []URL `xml:"url"`
}

// URL represents a single URL entry in a sitemap
type URL struct {
	Loc string `xml:"loc"`
}

// PageResult represents the complete analysis result for a single page
// including both mobile and desktop results
type PageResult struct {
	URL      string
	Mobile   Result
	Desktop  Result
	Duration time.Duration
}
