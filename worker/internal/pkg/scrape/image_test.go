package scrape

import (
	"github.com/nazar256/intopwa/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_detectProps(t *testing.T) {
	pngImage, err := os.ReadFile("tests/fixtures/apple-touch.png")
	require.NoError(t, err)

	icoImage, err := os.ReadFile("tests/fixtures/favicon.ico")
	require.NoError(t, err)

	tests := []struct {
		name          string
		image         []byte
		expectedProps domain.ImageProps
		wantErr       bool
	}{
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:  "png",
			image: pngImage,
			expectedProps: domain.ImageProps{
				MimeType: "image/png",
				Size: domain.ImageSize{
					Width:  160,
					Height: 160,
				},
			},
		},
		{
			name:  "ico",
			image: icoImage,
			expectedProps: domain.ImageProps{
				MimeType: "image/x-icon",
				Size: domain.ImageSize{
					Width:  48,
					Height: 48,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeImgProps(tt.image, "")
			assert.True(t, (err != nil) == tt.wantErr)

			assert.Equal(t, tt.expectedProps, got)
		})
	}
}
