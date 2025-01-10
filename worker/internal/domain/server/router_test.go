package server

import (
	"github.com/nazar256/intopwa/internal/domain"
	"github.com/nazar256/intopwa/internal/domain/server/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		name                string
		url                 string
		initMocks           func(fetcher *mocks.IconsFetcher)
		expectedStatus      int
		expectedContentType string
		expectedSubstrings  []string
	}{
		{
			name:                "App page",
			url:                 "/a/google.com/some/path/",
			expectedStatus:      http.StatusOK,
			expectedContentType: "text/html",
			expectedSubstrings: []string{
				"<title>App for google.com</title>",
				"/a/google.com/some/path/manifest.json",
				"/a/google.com/some/path/service-worker.js",
			},
			initMocks: func(fetcher *mocks.IconsFetcher) {
				u, _ := url.Parse("https://google.com/some/path/")
				var iconURLs []*url.URL
				fetcher.EXPECT().CacheIcons(mock.Anything, u, iconURLs).
					Return(nil).Once()
			},
		},
		{
			name: "Test valid icon path",
			url:  "/i/www.wikipedia.org/static/favicon.ico",
			initMocks: func(fetcher *mocks.IconsFetcher) {
				u, _ := url.Parse("https://www.wikipedia.org/static/favicon.ico")
				fetcher.EXPECT().One(mock.Anything, u).
					Return(
						domain.Icon{
							URL:  u,
							Body: []byte{},
							Props: domain.ImageProps{
								MimeType: "image/x-icon",
								Size: domain.ImageSize{
									Width:  64,
									Height: 64,
								},
							},
						},
						nil,
					).Once()
			},
			expectedContentType: "image/x-icon",
			expectedStatus:      http.StatusOK,
		},
		{
			name:           "Test invalid path",
			url:            "/invalid/",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "manifest",
			url:  "/a/www.wikipedia.org/manifest.json",
			initMocks: func(fetcher *mocks.IconsFetcher) {
				u, _ := url.Parse("https://www.wikipedia.org")
				iconU, _ := url.Parse("https://www.wikipedia.org/static/favicon.ico")
				fetcher.EXPECT().FetchIcons(mock.Anything, u).
					Return(
						[]domain.Icon{{
							URL:  iconU,
							Body: []byte{},
							Props: domain.ImageProps{
								MimeType: "image/x-icon",
								Size: domain.ImageSize{
									Width:  64,
									Height: 64,
								},
							},
						}},
					).Once()
			},
			expectedStatus:      http.StatusOK,
			expectedContentType: "application/json",
			expectedSubstrings: []string{
				"image/x-icon",
				"64x64",
				"/i/www.wikipedia.org/static/favicon.ico",
				"\"/a/www.wikipedia.org/redirect.html\"",
			},
		},
		{
			name:                "service workers",
			url:                 "/a/www.wikipedia.org/service-worker.js",
			expectedStatus:      http.StatusOK,
			expectedContentType: "application/javascript",
			expectedSubstrings: []string{
				"addEventListener",
			},
		},
		{
			name:                "redirect page",
			url:                 "/a/familylink.google.com/redirect.html",
			expectedStatus:      http.StatusOK,
			expectedContentType: "text/html",
			expectedSubstrings: []string{
				"<meta http-equiv=\"refresh\" content=\"0;url=https://familylink.google.com\">",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iconsFetcherMock := mocks.NewIconsFetcher(t)
			if tt.initMocks != nil {
				tt.initMocks(iconsFetcherMock)
			}

			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			fetcher := New(iconsFetcherMock)
			handler := fetcher.Router()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedContentType != "" {
				assert.Equal(t, tt.expectedContentType, rr.Header().Get("Content-Type"))
			}

			body := rr.Body.String()
			for _, expectedSubstr := range tt.expectedSubstrings {
				assert.Contains(t, body, expectedSubstr)
			}
		})
	}
}
