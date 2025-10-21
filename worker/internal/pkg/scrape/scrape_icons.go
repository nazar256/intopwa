package scrape

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/nazar256/intopwa/internal/domain"
	"golang.org/x/sync/errgroup"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

const (
	mobileUserAgent = "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Mobile Safari/537.36"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var iconSelectors = []string{
	"link[rel=icon]",
	"link[rel=apple-touch-icon]",
	"link[rel='shortcut icon']",
}

var tryIconURIs = []string{
	"/favicon.ico",
	"/favicon.svg",
}

type iconsScraper struct {
	client httpClient
}

func NewIconsScraper(client httpClient) *iconsScraper {
	if client == nil {
		panic("nil client passed")
	}
	return &iconsScraper{
		client: client,
	}
}

// scrapIcons scraps favicon, apple-touch and shortcut icons from the given URL using goquery
func (f *iconsScraper) ScrapeIconURLs(ctx context.Context, pageURL *url.URL) (iconURLs []*url.URL, err error) {
	// fetch the page
	doc, err := f.fetchPage(ctx, pageURL)
	if err != nil {
		return iconURLs, fmt.Errorf("failed to fetch page: %w", err)
	}

	for _, selector := range iconSelectors {
		scrapedIconURLs, err := f.scrapeIcons(doc, pageURL, selector)
		if err != nil {
			return iconURLs, fmt.Errorf("scrape icon error: %w", err)
		}
		iconURLs = append(iconURLs, scrapedIconURLs...)
	}

	for _, uri := range tryIconURIs {
		iconURLParts := []string{pageURL.Scheme, "://", pageURL.Hostname()}
		if pageURL.Port() != "" {
			iconURLParts = append(iconURLParts, ":", pageURL.Port())
		}
		iconURLParts = append(iconURLParts, uri)
		defaultIconURL, err := url.Parse(strings.Join(iconURLParts, ""))
		if err != nil {
			return iconURLs, fmt.Errorf("failed to parse default favicon URL: %w", err)
		}
		iconURLs = append(iconURLs, defaultIconURL)
	}

	slices.SortFunc(iconURLs, func(a, b *url.URL) int {
		return strings.Compare(a.String(), b.String())
	})
	iconURLs = slices.CompactFunc(iconURLs, func(a, b *url.URL) bool {
		return a.String() == b.String()
	})

	return iconURLs, nil
}

func (f *iconsScraper) fetchPage(ctx context.Context, url *url.URL) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", mobileUserAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return doc, nil
}

func (f *iconsScraper) scrapeIcons(
	doc *goquery.Document,
	url *url.URL,
	selector string,
) (iconURLs []*url.URL, err error) {
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		iconURL, parseErr := url.Parse(href)
		if parseErr != nil {
			err = fmt.Errorf("failed to parse icon URL (%s): %w", iconURL.String(), err)
			return
		}

		iconURLs = append(iconURLs, iconURL)
	})

	return
}

func (f *iconsScraper) DownloadIcons(ctx context.Context, iconURLs []*url.URL) (icons []domain.Icon, err error) {
	iconCh := make(chan domain.Icon, len(iconURLs))

	var fetchGroup, collectGroup errgroup.Group

	for _, u := range iconURLs {
		fetchGroup.Go(func() error {
			icon, err := f.downloadIcon(ctx, u)
			if err != nil {
				err = fmt.Errorf("failed to download icon: %w", err)
				if icon.Body == nil {
					return err
				}
				slog.Error("", "err", err)
			}
			iconCh <- icon
			return nil
		})
	}

	collectGroup.Go(func() error {
		for icon := range iconCh {
			icons = append(icons, icon)
		}
		return nil
	})

	_ = fetchGroup.Wait()
	close(iconCh)

	_ = collectGroup.Wait()

	return
}

func (f *iconsScraper) downloadIcon(ctx context.Context, iconURL *url.URL) (icon domain.Icon, err error) {
	icon.URL = iconURL

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, iconURL.String(), nil)
	if err != nil {
		return icon, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", mobileUserAgent)
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")

	resp, err := f.client.Do(req)
	if err != nil {
		return icon, fmt.Errorf("failed to fetch icon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return icon, fmt.Errorf("failed to fetch icon, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return icon, fmt.Errorf("failed to read icon body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		detectedContentType := http.DetectContentType(body)
		if strings.HasPrefix(detectedContentType, "image/") {
			contentType = detectedContentType
		} else {
			return icon, fmt.Errorf(
				"invalid icon content type: %s (detected: %s)",
				contentType,
				detectedContentType,
			)
		}
	}

	icon.Body = body

	props, err := decodeImgProps(body, contentType)
	if err != nil {
		return icon, fmt.Errorf("failed to decode image props: %w", err)
	}

	icon.Props = props

	return icon, nil
}
