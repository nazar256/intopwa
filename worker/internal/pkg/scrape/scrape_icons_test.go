package scrape_test

import (
	"context"
	"github.com/nazar256/intopwa/internal/domain"
	"github.com/nazar256/intopwa/internal/pkg/scrape"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestScrapIconURLs(t *testing.T) {
	const (
		noIconsURI = "/fixtures/no-icons.html"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			http.ServeFile(w, r, "./tests/fixtures/favicon.ico")
		} else {
			http.FileServer(http.Dir("./tests")).ServeHTTP(w, r)
		}
	}))
	defer server.Close()

	dummyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	tests := map[string]struct {
		uri           string
		expectedIcons []domain.Icon
		expectedURIs  []string
		//initMocks     func(kv *mocks.Kv)
	}{
		"no icons": {
			uri:          noIconsURI,
			expectedURIs: []string{"/favicon.ico", "/favicon.svg"},
		},
		"favicon": {
			uri:          "/fixtures/favicons-only.html",
			expectedURIs: []string{"/favicon.ico", "/favicon.svg"},
		},
		"multiple": {
			uri: "/fixtures/multiple-icons.html",
			expectedURIs: []string{
				"/favicon.ico",
				"/favicon.svg",
				"/fixtures/apple-touch.png",
				"/fixtures/favicon.ico",
			},
		},
		"epicgames": {
			uri: "/fixtures/epicgames.html",
			expectedURIs: []string{
				"/favicon.ico",
				"/favicon.svg",
				"/epic-store/static/favicon.ico",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			srv := server
			if test.uri == noIconsURI {
				srv = dummyServer
			}

			scraper := scrape.NewIconsScraper(srv.Client())

			u, _ := url.Parse(srv.URL + test.uri)
			iconURLs, err := scraper.ScrapeIconURLs(ctx, u)
			assert.NoError(t, err)

			var iconURLStrings []string
			for _, u := range iconURLs {
				iconURLStrings = append(
					iconURLStrings,
					strings.Join(
						[]string{
							u.Path,
							u.RawQuery,
						},
						"",
					),
				)
			}

			slices.Sort(iconURLStrings)
			slices.Sort(test.expectedURIs)

			assert.Equal(t, test.expectedURIs, iconURLStrings)
		})
	}
}
