package server

import (
	"bytes"
	"github.com/nazar256/intopwa/internal/domain"
	"github.com/nazar256/intopwa/internal/domain/server/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"mime/multipart"
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
				"/a/google.com/some/path/manifest.json?v=",
				"/a/google.com/some/path/service-worker.js",
				"/default-app-icon.png",
			},
			initMocks: func(fetcher *mocks.IconsFetcher) {
				u, _ := url.Parse("https://google.com/some/path/")
				var iconURLs []*url.URL
				fetcher.EXPECT().CacheIcons(mock.Anything, u, iconURLs).
					Return(nil).Once()
				fetcher.EXPECT().FetchIcons(mock.Anything, u).
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

func TestRouterUploadsIconFileForApp(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(uploadedIconFormField, "icon.png")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write(tinyPNG)
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	appU, _ := url.Parse("https://example.com")
	expectedIcon, err := readUploadedIcon(readSeekCloser{Reader: bytes.NewReader(tinyPNG)}, &multipart.FileHeader{
		Filename: "icon.png",
		Size:     int64(len(tinyPNG)),
		Header:   map[string][]string{"Content-Type": []string{"image/png"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	iconsFetcherMock := mocks.NewIconsFetcher(t)
	iconsFetcherMock.EXPECT().StoreUploadedIcon(mock.Anything, appU, mock.MatchedBy(func(icon domain.Icon) bool {
		return icon.URL.String() == expectedIcon.URL.String() && icon.Props.MimeType == "image/png" && icon.Props.Size.String() == "1x1"
	})).Return(nil).Once()
	iconsFetcherMock.EXPECT().CacheIcons(mock.Anything, appU, []*url.URL(nil)).Return(nil).Once()
	iconsFetcherMock.EXPECT().FetchIcons(mock.Anything, appU).Return([]domain.Icon{expectedIcon}).Once()

	req, err := http.NewRequest(http.MethodPost, "/a/example.com", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	New(iconsFetcherMock).Router().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/html", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Body.String(), "/i/"+domain.UploadedIconsHost+expectedIcon.URL.Path)
}

func TestRouterRejectsAmbiguousUploadedAndURLIcons(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("icons[]", "https://example.com/icon.png"); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile(uploadedIconFormField, "icon.png")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write(tinyPNG)
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/a/example.com", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	New(mocks.NewIconsFetcher(t)).Router().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Choose either icon URLs or one uploaded icon file")
}
