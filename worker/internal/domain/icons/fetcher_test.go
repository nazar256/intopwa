package icons

import (
	"context"
	"net/url"
	"testing"

	"github.com/nazar256/intopwa/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeScraper struct {
	downloadedURLs []*url.URL
}

func (f *fakeScraper) ScrapeIconURLs(_ context.Context, _ *url.URL) ([]*url.URL, error) {
	return nil, nil
}

func (f *fakeScraper) DownloadIcons(_ context.Context, iconURLs []*url.URL) ([]domain.Icon, error) {
	f.downloadedURLs = append(f.downloadedURLs, iconURLs...)
	return nil, nil
}

type fakeIconsCache struct {
	icons map[string]domain.Icon
}

func (f *fakeIconsCache) Store(icons []domain.Icon) error {
	if f.icons == nil {
		f.icons = map[string]domain.Icon{}
	}
	for _, icon := range icons {
		f.icons[icon.URL.String()] = icon
	}
	return nil
}

func (f *fakeIconsCache) Get(urls []*url.URL) ([]domain.Icon, bool, error) {
	icons := make([]domain.Icon, 0, len(urls))
	for _, u := range urls {
		icon, ok := f.icons[u.String()]
		if !ok {
			return nil, false, nil
		}
		icons = append(icons, icon)
	}
	return icons, true, nil
}

type fakeLinksCache struct {
	urls map[string][]*url.URL
}

func (f *fakeLinksCache) GetIconURLs(u *url.URL) ([]*url.URL, bool, error) {
	urls, ok := f.urls[u.String()]
	return urls, ok, nil
}

func (f *fakeLinksCache) StoreIconURLs(u *url.URL, iconsURLs []*url.URL) error {
	if f.urls == nil {
		f.urls = map[string][]*url.URL{}
	}
	f.urls[u.String()] = iconsURLs
	return nil
}

func uploadedIconFixture(t *testing.T) domain.Icon {
	t.Helper()
	iconURL, err := url.Parse("https://" + domain.UploadedIconsHost + "/abcdef.png")
	require.NoError(t, err)
	return domain.Icon{
		URL:   iconURL,
		Body:  []byte("png-bytes"),
		Props: domain.ImageProps{MimeType: "image/png"},
	}
}

func TestStoreUploadedIconPersistsBlobAndPageReference(t *testing.T) {
	iconsCache := &fakeIconsCache{}
	linksCache := &fakeLinksCache{}
	fetcher := NewIconsFetcher(&fakeScraper{}, iconsCache, linksCache)

	pageURL, err := url.Parse("https://example.com")
	require.NoError(t, err)
	icon := uploadedIconFixture(t)

	require.NoError(t, fetcher.StoreUploadedIcon(context.Background(), pageURL, icon))

	stored, found, err := iconsCache.Get([]*url.URL{icon.URL})
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, icon.Body, stored[0].Body)

	linked, found, err := linksCache.GetIconURLs(pageURL)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, []*url.URL{icon.URL}, linked)
}

func TestOneServesUploadedIconFromCacheWithoutNetwork(t *testing.T) {
	scraper := &fakeScraper{}
	iconsCache := &fakeIconsCache{}
	fetcher := NewIconsFetcher(scraper, iconsCache, &fakeLinksCache{})

	icon := uploadedIconFixture(t)
	require.NoError(t, iconsCache.Store([]domain.Icon{icon}))

	got, err := fetcher.One(context.Background(), icon.URL)
	require.NoError(t, err)
	assert.Equal(t, icon.Body, got.Body)
	assert.Empty(t, scraper.downloadedURLs)
}

func TestOneNeverDownloadsUploadedIconHost(t *testing.T) {
	scraper := &fakeScraper{}
	fetcher := NewIconsFetcher(scraper, &fakeIconsCache{}, &fakeLinksCache{})

	icon := uploadedIconFixture(t)
	_, err := fetcher.One(context.Background(), icon.URL)

	require.Error(t, err)
	assert.ErrorContains(t, err, "uploaded icon is not cached")
	assert.Empty(t, scraper.downloadedURLs)
}

func TestOneDownloadsOtherHostsOnCacheMiss(t *testing.T) {
	scraper := &fakeScraper{}
	fetcher := NewIconsFetcher(scraper, &fakeIconsCache{}, &fakeLinksCache{})

	iconURL, err := url.Parse("https://example.com/icon.png")
	require.NoError(t, err)

	_, err = fetcher.One(context.Background(), iconURL)
	require.NoError(t, err)
	assert.Equal(t, []*url.URL{iconURL}, scraper.downloadedURLs)
}

func TestStoreUploadedIconRejectsNilURL(t *testing.T) {
	fetcher := NewIconsFetcher(&fakeScraper{}, &fakeIconsCache{}, &fakeLinksCache{})

	pageURL, err := url.Parse("https://example.com")
	require.NoError(t, err)

	err = fetcher.StoreUploadedIcon(context.Background(), pageURL, domain.Icon{})
	require.Error(t, err)
}
