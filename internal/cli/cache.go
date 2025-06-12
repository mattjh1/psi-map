package cli

import (
	"fmt"
	"strings"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/utils"
	"github.com/urfave/cli/v2"
)

// Add cache management commands
func cacheCommands() *cli.Command {
	return &cli.Command{
		Name:  "cache",
		Usage: "Cache management commands",
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List cached results",
				Action: cacheListCommand,
			},
			{
				Name:   "clean",
				Usage:  "Remove expired cache files",
				Action: cacheCleanCommand,
			},
			{
				Name:   "clear",
				Usage:  "Clear all cached results",
				Action: cacheClearCommand,
			},
		},
	}
}

func cacheListCommand(c *cli.Context) error {
	ttl := c.Int("cache-ttl")
	cacheInfos, err := utils.ListCacheFiles(ttl)
	if err != nil {
		return err
	}

	if len(cacheInfos) == 0 {
		fmt.Println("No cached results found")
		return nil
	}

	fmt.Printf("Found %d cached result(s) (TTL: %dh):\n\n", len(cacheInfos), ttl)
	fmt.Printf("%-12s %-8s %-50s %s\n", "AGE", "STATUS", "SITEMAP", "HASH")
	fmt.Println(strings.Repeat("-", constants.SeparatorLength))

	for _, info := range cacheInfos {
		status := "VALID"
		if info.IsExpired {
			status = "EXPIRED"
		}

		sitemap := info.SitemapURL
		if len(sitemap) > 45 {
			sitemap = "..." + sitemap[len(sitemap)-42:]
		}

		fmt.Printf("%-12s %-8s %-50s %s\n",
			info.Age, status, sitemap, info.Hash[:8]+"...")
	}
	return nil
}

func cacheCleanCommand(c *cli.Context) error {
	ttl := c.Int("cache-ttl")
	cleaned, err := utils.CleanExpiredCache(ttl)
	if err != nil {
		return err
	}

	if cleaned == 0 {
		fmt.Printf("No expired cache files found (TTL: %dh)\n", ttl)
	} else {
		fmt.Printf("Cleaned %d expired cache file(s) (TTL: %dh)\n", cleaned, ttl)
	}
	return nil
}

func cacheClearCommand(c *cli.Context) error {
	if err := utils.ClearCache(); err != nil {
		return err
	}
	fmt.Println("Cache cleared successfully")
	return nil
}
