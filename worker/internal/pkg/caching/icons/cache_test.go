package icons

import (
	"github.com/nazar256/intopwa/internal/domain"
	"github.com/nazar256/intopwa/internal/pkg/caching/icons/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

const (
	googleFaviconUrlString  = "https://google.com/static/favicon.ico"
	googleFaviconOnlyRecord = `
			[{"URL":{"Scheme":"https","Opaque":"","User":null,"Host":"google.com",
			"Path":"/static/favicon.ico","RawPath":"","OmitHost":false,"ForceQuery":false,"RawQuery":"","Fragment":"",
			"RawFragment":""},"Body":"YQ==","Props":{"MimeType":"image/x-icon","Size":{"Width":64,"Height":64}}}]
`
)

var googleFaviconUrl, _ = url.Parse(googleFaviconUrlString)

func TestCache_Get(t *testing.T) {
	testCases := []struct {
		name          string
		urls          []string
		kv            map[string]string
		expectedIcons []domain.Icon
	}{
		{
			name: "single favicon",
			urls: []string{"https://google.com/static/favicon.ico"},
			kv: map[string]string{
				"google.com": googleFaviconOnlyRecord,
			},
			expectedIcons: []domain.Icon{
				{
					URL:  googleFaviconUrl,
					Body: []byte("a"),
					Props: domain.ImageProps{
						MimeType: "image/x-icon",
						Size:     domain.ImageSize{64, 64},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var urls []*url.URL
			for _, urlString := range tc.urls {
				u, err := url.Parse(urlString)
				require.NoError(t, err)
				urls = append(urls, u)
			}

			kvMock := mocks.NewKv(t)
			for key, val := range tc.kv {
				kvMock.EXPECT().Get(key).
					Return([]byte(val), nil).
					Once()
			}

			// Initialize your cache here
			c := NewCache(kvMock)

			icons, found, err := c.Get(urls)
			assert.NoError(t, err)
			assert.True(t, found)

			assert.Equal(t, tc.expectedIcons, icons)
		})
	}
}

func TestCache_Store(t *testing.T) {
	// Define your test cases here
	testCases := []struct {
		name            string
		url             string
		icons           []domain.Icon
		initMocks       func(kv *mocks.Kv)
		expectedKvValue string
	}{
		{
			name:  "empty list NOP",
			url:   googleFaviconUrlString,
			icons: []domain.Icon{},
			initMocks: func(kv *mocks.Kv) {

			},
		},
		{
			name: "favicon only",
			url:  googleFaviconUrlString,
			icons: []domain.Icon{
				{
					URL:  googleFaviconUrl,
					Body: []byte("a"),
					Props: domain.ImageProps{
						MimeType: "image/x-icon",
						Size:     domain.ImageSize{64, 64},
					},
				},
			},
			initMocks: func(kv *mocks.Kv) {
				kv.EXPECT().Get("google.com").
					Return(nil, nil).Once()
				kv.EXPECT().Put("google.com", mock.MatchedBy(func(b []byte) bool {
					return assert.JSONEq(t, googleFaviconOnlyRecord, string(b))
				})).Return(nil).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kvMock := mocks.NewKv(t)
			defer func() {
				kvMock.ExpectedCalls = nil
			}()

			if tc.initMocks != nil {
				tc.initMocks(kvMock)
			}

			// Initialize your cache here
			c := NewCache(kvMock)

			err := c.Store(tc.icons)
			assert.NoError(t, err)
		})
	}
}
