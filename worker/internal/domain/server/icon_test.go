package server

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestParseIconURL(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *url.URL
		err      error
	}{
		{
			name:     "Valid URL",
			input:    "https://example.com/i/google.com/static/icon.png",
			expected: &url.URL{Scheme: "https", Host: "google.com", Path: "/static/icon.png"},
			err:      nil,
		},
		{
			name:     "Invalid URL",
			input:    "https://example.com/google.com",
			expected: nil,
			err:      errors.New("Invalid request format"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, _ := url.Parse(tc.input)
			iconU, err := parseIconURL(u)

			if tc.err != nil {
				assert.Error(t, err)
			}
			assert.Equal(t, tc.expected, iconU)
		})
	}
}
