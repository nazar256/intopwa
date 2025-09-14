package server

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestParseAppURL(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *appURL
		err      error
	}{
		{
			name:     "Valid URL",
			input:    "https://example.com/a/google.com/some/path",
			expected: &appURL{URL: url.URL{Scheme: "https", Host: "google.com", Path: "/some/path"}},
			err:      nil,
		},
		{
			name:     "Manifest URL",
			input:    "https://example.com/a/google.com/some/path" + manifestPath,
			expected: &appURL{URL: url.URL{Scheme: "https", Host: "google.com", Path: "/some/path"}},
			err:      nil,
		},
		{
			name:     "URL with query",
			input:    "https://example.com/a/google.com/some/path?foo=bar",
			expected: &appURL{URL: url.URL{Scheme: "https", Host: "google.com", Path: "/some/path", RawQuery: "foo=bar"}},
			err:      nil,
		},
		{
			name:     "Invalid URL",
			input:    "https://example.com/google.com",
			expected: nil,
			err:      fmt.Errorf("failed to parse app URL: parse \"https://invalid\": invalid URI for request"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, _ := url.Parse(tc.input)
			appU, err := parseAppURL(u)

			if tc.err != nil {
				assert.Error(t, err)
			}
			assert.Equal(t, tc.expected, appU)
		})
	}
}
