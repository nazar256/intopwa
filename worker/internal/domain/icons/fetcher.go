package icons

import (
	"cmp"
	"context"
	"fmt"
	"github.com/nazar256/intopwa/internal/domain"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

type scraper interface {
	ScrapeIconURLs(ctx context.Context, u *url.URL) ([]*url.URL, error)
	DownloadIcons(ctx context.Context, iconURLs []*url.URL) (icons []domain.Icon, err error)
}

type iconsCache interface {
	Store(icons []domain.Icon) error
	Get(urls []*url.URL) ([]domain.Icon, bool, error)
}

type linksCache interface {
	GetIconURLs(u *url.URL) (urls []*url.URL, found bool, err error)
	StoreIconURLs(u *url.URL, iconsURLs []*url.URL) error
}

type fetcher struct {
	scraper    scraper
	iconsCache iconsCache
	linksCache linksCache
}

func NewIconsFetcher(s scraper, icons iconsCache, links linksCache) *fetcher {
	return &fetcher{
		scraper:    s,
		iconsCache: icons,
		linksCache: links,
	}
}

func (f *fetcher) CacheIcons(ctx context.Context, pageURL *url.URL, iconURLs []*url.URL) (err error) {
	if len(iconURLs) == 0 {
		return nil
	}

	icons, err := f.scraper.DownloadIcons(ctx, iconURLs)
	if err != nil {
		return fmt.Errorf("failed to download icons: %w", err)
	}

	if len(icons) == 0 {
		return nil
	}

	err = f.iconsCache.Store(icons)
	if err != nil {
		return fmt.Errorf("failed to store icons: %w", err)
	}

	err = f.linksCache.StoreIconURLs(pageURL, iconURLs)
	if err != nil {
		slog.Error("failed to store icon URLs", "err", err)
		return nil
	}

	return nil
}

func (f *fetcher) FetchIcons(ctx context.Context, u *url.URL) []domain.Icon {
	iconURLs, iconsURLsFound, err := f.linksCache.GetIconURLs(u)
	if err != nil {
		slog.Error("failed to read icon URLs from cache", "err", err)
		return nil
	}

	if !iconsURLsFound {
		iconURLs, err = f.scraper.ScrapeIconURLs(ctx, u)
		if err != nil {
			slog.Error("failed to scrape icon URLs", "err", err)
			return nil
		}

		err = f.linksCache.StoreIconURLs(u, iconURLs)
		if err != nil {
			slog.Error("failed to store icon URLs", "err", err)
			return nil
		}
	}

	if len(iconURLs) == 0 {
		return nil
	}

	icons, iconsFound, err := f.iconsCache.Get(iconURLs)
	if err != nil {
		slog.Error("failed to read icons from cache", "err", err)
		return nil
	}

	if !iconsFound {
		icons, err = f.scraper.DownloadIcons(ctx, iconURLs)
		if err != nil {
			slog.Error("failed to download icons", "err", err)
			return icons
		}

		err = f.iconsCache.Store(icons)
		if err != nil {
			slog.Error("failed to store icons in cache", "err", err)
			return icons
		}
	}

	return ensureBigIcon(icons)
}

func (f *fetcher) One(ctx context.Context, iconURL *url.URL) (domain.Icon, error) {
	icons, found, err := f.iconsCache.Get([]*url.URL{iconURL})
	if err != nil {
		return domain.Icon{}, fmt.Errorf("failed to read icon from iconsCache: %w", err)
	}

	if !found {
		icons, err = f.scraper.DownloadIcons(ctx, []*url.URL{iconURL})
		if err != nil {
			return domain.Icon{}, fmt.Errorf("failed to download Icon: %w", err)
		}

		err = f.iconsCache.Store(icons)
		if err != nil {
			return domain.Icon{}, fmt.Errorf("failed to store icon: %w", err)
		}

	}

	if len(icons) == 0 {
		return domain.Icon{}, nil
	}

	return icons[0], nil
}

func ensureBigIcon(icons []domain.Icon) []domain.Icon {
	if len(icons) == 0 {
		return icons
	}

	normalized := make([]domain.Icon, 0, len(icons))
	for _, icon := range icons {
		normalized = append(normalized, normalizeIcon(icon))
	}

	if containsExactSizeIcon(normalized, 512, 512) {
		return normalized
	}

	resizingCandidate := normalizeIcon(pickResizingCandidate(normalized))

	return append(normalized, domain.Icon{
		URL:  resizingCandidate.URL,
		Body: resizingCandidate.Body,
		Props: domain.ImageProps{
			MimeType: resizingCandidate.Props.MimeType,
			Size: domain.ImageSize{
				Width:  512,
				Height: 512,
			},
		},
	})
}

func normalizeIcon(icon domain.Icon) domain.Icon {
	if icon.Props.MimeType == "" && len(icon.Body) > 0 {
		icon.Props.MimeType = http.DetectContentType(icon.Body)
	}
	return icon
}

func containsExactSizeIcon(icons []domain.Icon, width, height int) bool {
	for _, icon := range icons {
		if icon.Props.Size.Width == width && icon.Props.Size.Height == height {
			return true
		}
	}
	return false
}

func pickResizingCandidate(icons []domain.Icon) (candidate domain.Icon) {
	if len(icons) == 0 {
		return
	}

	for _, icon := range icons {
		if strings.Contains(icon.Props.MimeType, "svg") {
			return icon
		}
	}

	iconsToSort := make([]domain.Icon, len(icons))
	copy(iconsToSort, icons)

	slices.SortFunc(iconsToSort, func(a, b domain.Icon) int {
		pixelsA := a.Props.Size.Width * a.Props.Size.Height
		pixelsB := b.Props.Size.Width * b.Props.Size.Height
		return cmp.Compare(pixelsA, pixelsB)
	})

	return iconsToSort[0]
}
