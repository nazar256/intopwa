//go:build cloudflare
// +build cloudflare

package main

import (
	"fmt"
	"github.com/nazar256/intopwa/internal/domain/icons"
	"github.com/nazar256/intopwa/internal/domain/server"
	cache_icons "github.com/nazar256/intopwa/internal/pkg/caching/icons"
	"github.com/nazar256/intopwa/internal/pkg/caching/links"
	"github.com/nazar256/intopwa/internal/pkg/scrape"
	compat_cf "github.com/nazar256/intopwa/pkg/compatibility/cloudflare"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"
	"log/slog"
	"os"
	"time"
)

const (
	iconsKVNamespace = "ICONS"
	linksKVNamespace = "LINKS"
	thirtyDays       = 30 * 24 * time.Hour
)

func main() {
	// initialize KV namespace instance
	_, err := cloudflare.NewKVNamespace(iconsKVNamespace)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to init KV: %v", err)
		os.Exit(1)
	}

	scraper := scrape.NewIconsScraper(compat_cf.NewFetcher())

	iconsKV, err := cloudflare.NewKVNamespace(iconsKVNamespace)
	if err != nil {
		slog.Error("failed to initialize KV", "err", err, "namespace", iconsKVNamespace)
		os.Exit(1)
	}

	linksKV, err := cloudflare.NewKVNamespace(linksKVNamespace)
	if err != nil {
		slog.Error("failed to initialize KV", "err", err, "namespace", iconsKVNamespace)
		os.Exit(1)
	}

	iconsKVWrapper := compat_cf.NewKV(iconsKV, thirtyDays)
	linksKVWrapper := compat_cf.NewKV(linksKV, thirtyDays)

	iconsCache := cache_icons.NewCache(iconsKVWrapper)
	linksCache := links.NewCache(linksKVWrapper)

	fetcher := icons.NewIconsFetcher(scraper, iconsCache, linksCache)

	srv := server.New(fetcher)
	workers.Serve(srv.Router())
}
