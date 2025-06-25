package cli

import (
	"fmt"
	"math"
	"strings"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/logger"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

func cacheCommands() *cli.Command {
	return &cli.Command{
		Name:  "cache",
		Usage: "Manage cached PageSpeed Insights results",
		Description: `Manage cached results to optimize performance and storage.
        
Examples:
  psi-map cache list
  psi-map cache list --verbose
  psi-map cache clean --dry-run
  psi-map cache clear --force`,
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List cached results with details",
				Action: cacheListCommand,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show verbose cache information including sizes",
					},
					&cli.IntFlag{
						Name:  "cache-ttl",
						Value: constants.DefaultTTLHours,
						Usage: "Cache TTL in hours for status calculation",
					},
				},
			},
			{
				Name:   "clean",
				Usage:  "Remove expired cache files",
				Action: cacheCleanCommand,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "Show what would be deleted without actually deleting",
					},
					&cli.IntFlag{
						Name:  "cache-ttl",
						Value: constants.DefaultTTLHours,
						Usage: "Cache TTL in hours",
					},
				},
			},
			{
				Name:   "clear",
				Usage:  "Clear all cached results",
				Action: cacheClearCommand,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Force clear without confirmation prompt",
					},
				},
			},
		},
	}
}

func cacheListCommand(c *cli.Context) error {
	l := logger.GetLogger()
	u := l.UI(logger.WithUIStyle(&logger.UIStyle{
		TableBorderStyle: pterm.NewStyle(pterm.FgLightBlue),
		HeaderBgColor:    pterm.BgBlue,
	}))
	u.Clear()

	ttl := c.Int("cache-ttl")
	verbose := c.Bool("verbose")

	cacheInfos, err := utils.ListCacheFiles(ttl, true) // Always get full details
	if err != nil {
		l.Error("Failed to list cache files: %v", err)
		return fmt.Errorf("failed to list cache files: %w", err)
	}

	if len(cacheInfos) == 0 {
		l.Info("No cached results found")
		return nil
	}

	u.Header("Cached Sitemaps")
	l.Tagged("CACHE", "Found %d cached sitemap(s) (TTL: %dh)", "", len(cacheInfos), ttl)

	headers := []string{"TYPE", "AGE", "STATUS", "URL/SITEMAP", "HASH/SCORE"}
	if verbose {
		headers = append(headers, "SIZE")
	}

	data := make([][]string, 0)
	totalURLs, totalSize := 0, int64(0)

	for _, info := range cacheInfos {
		status := "VALID"
		if info.IsExpired {
			status = "EXPIRED"
		} else if info.StaleCount > 0 {
			status = "MIXED"
		}

		sitemap := truncateURL(info.SitemapURL, 60)
		urlsInfo := fmt.Sprintf("(%d URLs: %d valid, %d stale, %d expired)",
			info.URLCount, info.ValidCount, info.StaleCount, info.ExpiredCount)

		sitemapRow := []string{
			"📊 SITEMAP",
			info.Age,
			status,
			fmt.Sprintf("%s %s", sitemap, urlsInfo),
			info.Hash + "...",
		}

		if verbose {
			sitemapRow = append(sitemapRow, formatBytes(info.TotalSize))
		}
		data = append(data, sitemapRow)

		urlDetails, err := utils.GetURLCacheDetails(info.FullHash, ttl)
		if err == nil {
			for _, urlInfo := range urlDetails {
				urlStatus := getStatusIcon(urlInfo)
				url := truncateURL(urlInfo.URL, 70)
				scoreStr := ""
				if urlInfo.PerformanceScore > 0 {
					scoreStr = fmt.Sprintf("%d", int(math.Round(urlInfo.PerformanceScore)))
				}

				urlRow := []string{
					"  └─ URL",
					urlInfo.Age,
					urlStatus,
					url,
					scoreStr,
				}
				if verbose {
					urlRow = append(urlRow, formatBytes(urlInfo.CacheSize))
				}
				data = append(data, urlRow)
			}
		}

		if len(cacheInfos) > 1 {
			data = append(data, make([]string, len(headers)))
		}

		totalURLs += info.URLCount
		totalSize += info.TotalSize
	}

	if len(data) > 0 && isEmptyRow(data[len(data)-1]) {
		data = data[:len(data)-1]
	}

	u.Table(headers, data)

	if verbose {
		l.Info("Total: %d URLs, %s cache size", totalURLs, formatBytes(totalSize))
	}

	return nil
}

func cacheCleanCommand(c *cli.Context) error {
	l := logger.GetLogger()
	u := l.UI(logger.WithUIStyle(&logger.UIStyle{
		HeaderBgColor: pterm.BgYellow,
	}))
	u.Header("Cache Cleanup")

	ttl := c.Int("cache-ttl")
	dryRun := c.Bool("dry-run")

	l.Tagged("CACHE", "Starting cache cleanup (TTL: %dh)", "🧹", ttl)

	cleanedCount, err := utils.CleanExpiredCacheFiles(ttl, dryRun)
	if err != nil {
		l.Error("Cache cleanup failed: %v", err)
		return fmt.Errorf("failed to clean expired cache files: %w", err)
	}

	if cleanedCount == 0 {
		l.Success("No expired cache files found")
	} else {
		action := "would be removed"
		if !dryRun {
			action = "removed"
		}
		l.Success("Cache cleanup completed: %d expired file(s) %s", cleanedCount, action)
	}

	return nil
}

func cacheClearCommand(c *cli.Context) error {
	l := logger.GetLogger()
	u := l.UI(logger.WithUIStyle(&logger.UIStyle{
		HeaderBgColor: pterm.BgRed,
	}))
	u.Header("Cache Clear")

	force := c.Bool("force")
	if !force {
		response, err := u.Prompt("Are you sure you want to clear all cache data? This cannot be undone.", logger.ConfirmInput)
		if err != nil {
			return fmt.Errorf("failed to prompt for confirmation: %w", err)
		}
		confirmed, ok := response.(bool)
		if !ok {
			l.Error("unexpected type from confirmation prompt")
		}
		if !confirmed {
			l.Info("Cache clear canceled")
			return nil
		}
	}

	l.Tagged("CACHE", "Clearing all cache data", "🗑️")

	clearedCount, err := utils.ClearAllCacheFiles()
	if err != nil {
		l.Error("Failed to clear cache: %v", err)
		return fmt.Errorf("failed  to clear cache: %w", err)
	}

	if clearedCount == 0 {
		l.Info("No cache files found to clear")
	} else {
		l.Success("All cache data cleared: %d file(s) removed", clearedCount)
	}

	return nil
}

// Helper functions
func getStatusIcon(detail types.URLCacheDetail) string {
	if detail.IsExpired {
		return "EXP"
	} else if detail.IsStale {
		return " STL"
	} else if detail.HasErrors {
		return "ERR"
	}
	return "OK"
}

func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	if maxLen < 10 {
		return url[:maxLen]
	}
	if strings.HasPrefix(url, "http") {
		parts := strings.SplitN(url, "/", 4)
		if len(parts) >= 3 {
			domain := parts[0] + "//" + parts[2]
			if len(parts) == 4 {
				remaining := maxLen - len(domain) - 4
				if remaining > 0 && len(parts[3]) > remaining {
					return domain + "/..." + parts[3][len(parts[3])-remaining:]
				}
			}
		}
	}
	mid := maxLen - 6
	return url[:mid/2] + "..." + url[len(url)-mid/2:]
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if cell != "" {
			return false
		}
	}
	return true
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
