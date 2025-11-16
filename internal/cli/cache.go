package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattjh1/psi-map/internal/cache"
	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/logger"
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
  psi-map cache clean --dry-run
  psi-map cache clear --force`,
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List cached sitemap indexes",
				Action: cacheListCommand,
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

	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %w", err)
	}
	store, err := cache.NewFilesystemCacheStore(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to create cache store: %w", err)
	}
	defer store.Close()

	indexFiles, err := store.ListSitemapIndexFiles()
	if err != nil {
		return fmt.Errorf("failed to list cache indexes: %w", err)
	}

	if len(indexFiles) == 0 {
		l.Info("No cached results found")
		return nil
	}

	u.Header("Cached Sitemap Indexes")
	l.Tagged("CACHE", "Found %d cached sitemap index(es)", "", len(indexFiles))

	headers := []string{"Sitemap URL", "URL Count", "Last Updated", "Hash"}
	data := make([][]string, 0)

	for _, filename := range indexFiles {
		index, err := store.LoadSitemapIndexFile(filename)
		if err != nil {
			l.Warn("Failed to load index file %s: %v", filename, err)
			continue
		}
		row := []string{
			truncateURL(index.SitemapURL, 60),
			fmt.Sprintf("%d", len(index.URLs)),
			index.UpdatedAt.Format(time.RFC822),
			index.Hash,
		}
		data = append(data, row)
	}

	u.Table(headers, data)
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

	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %w", err)
	}
	store, err := cache.NewFilesystemCacheStore(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to create cache store: %w", err)
	}
	defer store.Close()

	cleanedCount, err := cache.CleanExpired(store, time.Duration(ttl)*time.Hour, dryRun)
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

	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		l.Error("Failed to get cache directory: %v", err)
		return fmt.Errorf("failed to get cache directory: %w", err)
	}

	if err := os.RemoveAll(cacheDir); err != nil {
		l.Error("Failed to clear cache: %v", err)
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	l.Success("All cache data cleared.")

	return nil
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
