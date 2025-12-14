//go:build integration
// +build integration

package server_test

import (
	"bytes"
	"encoding/json"
	"github.com/nazar256/intopwa/internal/domain/icons"
	"github.com/nazar256/intopwa/internal/domain/server"
	cache_icons "github.com/nazar256/intopwa/internal/pkg/caching/icons"
	"github.com/nazar256/intopwa/internal/pkg/caching/links"
	"github.com/nazar256/intopwa/internal/pkg/scrape"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
)

type PwaManifest struct {
	Name            string    `json:"name"`
	ShortName       string    `json:"short_name"`
	Icons           []PwaIcon `json:"icons,omitempty"`
	StartURL        string    `json:"start_url"`
	BackgroundColor string    `json:"background_color"`
	ThemeColor      string    `json:"theme_color"`
	Display         string    `json:"display"`
}

type PwaIcon struct {
	Src   string `json:"src"`
	Type  string `json:"type"`
	Sizes string `json:"sizes"`
}

func TestServer(t *testing.T) {
	iconUrls := []string{
		"https://static-assets-prod.epicgames.com/epic-store/static/favicon.ico",
		"https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/epicgames.svg",
	}

	values := url.Values{}
	for _, iconUrl := range iconUrls {
		values.Add("icons[]", iconUrl)
	}
	customIconsBody := []byte(values.Encode())

	udemyIconURLs := []string{
		"https://raw.githubusercontent.com/edent/SuperTinyIcons/master/images/svg/udemy.svg",
	}
	udemyIconValues := url.Values{}
	for _, iconUrl := range udemyIconURLs {
		udemyIconValues.Add("icons[]", iconUrl)
	}
	udemyIconsBody := []byte(udemyIconValues.Encode())

	shazamIconURLs := []string{
		"https://cdn-icons-png.flaticon.com/256/732/732242.png",
	}
	shazamIconValues := url.Values{}
	for _, iconUrl := range shazamIconURLs {
		shazamIconValues.Add("icons[]", iconUrl)
	}
	shazamIconsBody := []byte(shazamIconValues.Encode())

	chatgptIconURLs := []string{
		"https://pngimg.com/d/chatgpt_PNG14.png",
	}
	chatgptIconValues := url.Values{}
	for _, iconUrl := range chatgptIconURLs {
		chatgptIconValues.Add("icons[]", iconUrl)
	}
	chatgptIconsBody := []byte(chatgptIconValues.Encode())

	pixabayIconURLs := []string{
		"https://cdn.pixabay.com/photo/2023/05/08/00/43/chatgpt-7977357_1280.png",
	}
	pixabayIconValues := url.Values{}
	for _, iconUrl := range pixabayIconURLs {
		pixabayIconValues.Add("icons[]", iconUrl)
	}
	pixabayIconsBody := []byte(pixabayIconValues.Encode())

	seekLogoIconURLs := []string{
		"https://images.seeklogo.com/logo-png/46/1/chatgpt-logo-png_seeklogo-465219.png",
	}
	seekLogoIconValues := url.Values{}
	for _, iconUrl := range seekLogoIconURLs {
		seekLogoIconValues.Add("icons[]", iconUrl)
	}
	seekLogoIconsBody := []byte(seekLogoIconValues.Encode())

	tests := map[string]struct {
		uri              string
		postBody         []byte
		expectedTitle    string
		expectedManifest PwaManifest
	}{
		"Family Link": {
			uri:           "/a/familylink.google.com",
			expectedTitle: "App for familylink.google.com",
			expectedManifest: PwaManifest{
				Name: "familylink.google.com",
				Icons: []PwaIcon{
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_40.png",
						Type:  "image/png",
						Sizes: "40x40",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_192.png",
						Type:  "image/png",
						Sizes: "192x192",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_512.png",
						Type:  "image/png",
						Sizes: "512x512",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_60.png",
						Type:  "image/png",
						Sizes: "60x60",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_152.png",
						Type:  "image/png",
						Sizes: "152x152",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_87.png",
						Type:  "image/png",
						Sizes: "87x87",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_180.png",
						Type:  "image/png",
						Sizes: "180x180",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_58.png",
						Type:  "image/png",
						Sizes: "58x58",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_167.png",
						Type:  "image/png",
						Sizes: "167x167",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_1024.png",
						Type:  "image/png",
						Sizes: "1024x1024",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_favicon.ico",
						Type:  "image/x-icon",
						Sizes: "16x16",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_120.png",
						Type:  "image/png",
						Sizes: "120x120",
					},
					{
						Src:   "/i/www.gstatic.com/family/familylink/family_link_80.png",
						Type:  "image/png",
						Sizes: "80x80",
					},
				},
			},
		}, "Firefox Relay": {
			uri:           "/a/relay.firefox.com/accounts/profile",
			expectedTitle: "App for relay.firefox.com",
			expectedManifest: PwaManifest{
				Name: "relay.firefox.com/accounts/profile",
				Icons: []PwaIcon{
					{
						Src:   "/i/relay.firefox.com/favicon.svg",
						Type:  "image/svg+xml",
						Sizes: "512x512",
					},
					{
						Src:   "/i/relay.firefox.com/favicon.svg",
						Type:  "image/svg+xml",
						Sizes: "70x77",
					},
				},
			},
		},
		"Epic Games free": {
			uri:           "/a/store.epicgames.com/ru/free-games",
			postBody:      customIconsBody,
			expectedTitle: "App for store.epicgames.com",
			expectedManifest: PwaManifest{
				Name: "store.epicgames.com/ru/free-games",
				Icons: []PwaIcon{
					{
						Src:   "/i/raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/epicgames.svg",
						Type:  "image/svg+xml",
						Sizes: "24x24",
					},
					{
						Src:   "/i/raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/epicgames.svg",
						Type:  "image/svg+xml",
						Sizes: "512x512",
					},
					{
						Src:   "/i/static-assets-prod.epicgames.com/epic-store/static/favicon.ico",
						Type:  "image/vnd.microsoft.icon",
						Sizes: "48x48",
					},
				},
			},
		},
		"Udemy custom icon": {
			uri:           "/a/www.udemy.com/",
			postBody:      udemyIconsBody,
			expectedTitle: "App for www.udemy.com",
			expectedManifest: PwaManifest{
				Name: "www.udemy.com",
				Icons: []PwaIcon{
					{
						Src:   "/i/raw.githubusercontent.com/edent/SuperTinyIcons/master/images/svg/udemy.svg",
						Type:  "image/svg+xml",
						Sizes: "512x512",
					},
				},
			},
		},
		"Shazam custom icon": {
			uri:           "/a/www.shazam.com/",
			postBody:      shazamIconsBody,
			expectedTitle: "App for www.shazam.com",
			expectedManifest: PwaManifest{
				Name: "www.shazam.com",
				Icons: []PwaIcon{
					{
						Src:   "/i/cdn-icons-png.flaticon.com/256/732/732242.png",
						Type:  "image/png",
						Sizes: "256x256",
					},
					{
						Src:   "/i/cdn-icons-png.flaticon.com/256/732/732242.png",
						Type:  "image/png",
						Sizes: "512x512",
					},
				},
			},
		},
		"ChatGPT custom icon": {
			uri:           "/a/chatgpt.com/",
			postBody:      chatgptIconsBody,
			expectedTitle: "App for chatgpt.com",
			expectedManifest: PwaManifest{
				Name: "chatgpt.com",
				Icons: []PwaIcon{
					{
						Src:   "/i/pngimg.com/d/chatgpt_PNG14.png",
						Type:  "image/png",
						Sizes: "512x512",
					},
				},
			},
		},
		"ChatGPT pixabay icon": {
			uri:           "/a/chatgpt.com/",
			postBody:      pixabayIconsBody,
			expectedTitle: "App for chatgpt.com",
			expectedManifest: PwaManifest{
				Name: "chatgpt.com",
				Icons: []PwaIcon{
					{
						Src:   "/i/cdn.pixabay.com/photo/2023/05/08/00/43/chatgpt-7977357_1280.png",
						Type:  "image/png",
						Sizes: "1279x1280",
					},
					{
						Src:   "/i/cdn.pixabay.com/photo/2023/05/08/00/43/chatgpt-7977357_1280.png",
						Type:  "image/png",
						Sizes: "512x512",
					},
				},
			},
		},
		"ChatGPT seeklogo icon": {
			uri:           "/a/chatgpt.com/",
			postBody:      seekLogoIconsBody,
			expectedTitle: "App for chatgpt.com",
			expectedManifest: PwaManifest{
				Name: "chatgpt.com",
				Icons: []PwaIcon{
					{
						Src:   "/i/images.seeklogo.com/logo-png/46/1/chatgpt-logo-png_seeklogo-465219.png",
						Type:  "image/png",
						Sizes: "600x600",
					},
					{
						Src:   "/i/images.seeklogo.com/logo-png/46/1/chatgpt-logo-png_seeklogo-465219.png",
						Type:  "image/png",
						Sizes: "512x512",
					},
				},
			},
		},
	}

	scraper := scrape.NewIconsScraper(new(http.Client))
	iconsCache := cache_icons.NewCache(newMemKV())
	assetsCache := links.NewCache(newMemKV())
	iconsFetcher := icons.NewIconsFetcher(scraper, iconsCache, assetsCache)
	app := server.New(iconsFetcher)

	srv := httptest.NewServer(app.Router())

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			appPageResp := requestURI(t, srv, tt.uri, tt.postBody)
			manifestUrl, serviceWorkerURL := assertAppPage(t, appPageResp, tt.expectedTitle)

			serviceWorkerResp := requestURI(t, srv, serviceWorkerURL.String(), nil)
			assertServiceWorker(t, serviceWorkerResp)

			manifestResp := requestURI(t, srv, manifestUrl.String(), nil)
			iconsURLs := assertManifest(t, manifestResp, tt.expectedManifest)

			for _, iconURL := range iconsURLs {
				iconResp := requestURI(t, srv, iconURL.String(), nil)
				assertIcon(t, tt.expectedManifest, iconResp)
			}

			// Check consistency after caching
			appPageResp = requestURI(t, srv, tt.uri, tt.postBody)
			manifestUrl, _ = assertAppPage(t, appPageResp, tt.expectedTitle)

			manifestResp = requestURI(t, srv, manifestUrl.String(), nil)
			_ = assertManifest(t, manifestResp, tt.expectedManifest)
		})
	}
}

func requestURI(t *testing.T, srv *httptest.Server, uri string, postBody []byte) (resp *http.Response) {
	var err error

	switch {
	case postBody != nil:
		resp, err = srv.Client().Post(srv.URL+uri, "application/x-www-form-urlencoded", bytes.NewBuffer(postBody))
		if err != nil {
			t.Fatal(err)
		}
	default:
		resp, err = srv.Client().Get(srv.URL + uri)
		if err != nil {
			t.Fatal(err)
		}
	}

	return resp
}

func assertAppPage(t *testing.T, resp *http.Response, expectedTitle string) (manifestUrl *url.URL, serviceWorkerURL *url.URL) {
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Contains(t, string(body), expectedTitle)

	// manifest Regexp
	manifestRegex := regexp.MustCompile(`<link rel="manifest" href="(.+?)">`)
	matches := manifestRegex.FindStringSubmatch(string(body))
	require.Len(t, matches, 2)
	manifestUrlStr := matches[1]

	manifestUrl, err = url.Parse(manifestUrlStr)
	require.NoError(t, err)

	serviceWorkerRegex := regexp.MustCompile(`<script>\s+if \('serviceWorker' in navigator\) {\s+navigator.serviceWorker.register\('(.+?)'\)`)
	matches = serviceWorkerRegex.FindStringSubmatch(string(body))
	require.Len(t, matches, 2)
	serviceWorkerURLStr := matches[1]

	serviceWorkerURL, err = url.Parse(serviceWorkerURLStr)
	require.NoError(t, err)

	return manifestUrl, serviceWorkerURL
}

func assertServiceWorker(t *testing.T, resp *http.Response) {
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/javascript", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Contains(t, string(body), "addEventListener")
}

func assertManifest(t *testing.T, resp *http.Response, expectedManifest PwaManifest) (iconsURLs []*url.URL) {
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var manifest PwaManifest
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&manifest)
	require.NoError(t, err)
	require.NotNil(t, manifest)

	assert.Equal(t, expectedManifest.Name, manifest.Name)

	require.LessOrEqual(t, len(manifest.Icons), len(expectedManifest.Icons))

	for _, expectedIcon := range expectedManifest.Icons {
		icon := findIcon(manifest.Icons, expectedIcon.Src, expectedIcon.Sizes)
		require.NotNil(t, icon)
		assert.Equal(t, expectedIcon.Src, icon.Src)
		assert.Equal(t, expectedIcon.Type, icon.Type)
		assert.Equal(t, expectedIcon.Sizes, icon.Sizes)

		iconURL, err := url.Parse(icon.Src)
		require.NoError(t, err)
		iconsURLs = append(iconsURLs, iconURL)
	}

	return iconsURLs
}

func findIcon(icons []PwaIcon, src, sizes string) *PwaIcon {
	for _, icon := range icons {
		if icon.Src == src && (icon.Sizes == sizes || sizes == "") {
			return &icon
		}
	}
	return nil
}

func assertIcon(t *testing.T, expectedManifest PwaManifest, resp *http.Response) {
	icon := findIcon(expectedManifest.Icons, resp.Request.URL.Path, "")
	if icon == nil {
		return
	}

	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, icon.Type, resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.NotEmpty(t, body)
}

type memKV struct {
	icons map[string][]byte
}

func newMemKV() *memKV {
	return &memKV{
		icons: make(map[string][]byte),
	}
}

func (m *memKV) Get(key string) ([]byte, error) {
	value, ok := m.icons[key]
	if !ok {
		return nil, nil
	}
	return value, nil
}

func (m *memKV) Put(key string, value []byte) error {

	m.icons[key] = value
	return nil
}
