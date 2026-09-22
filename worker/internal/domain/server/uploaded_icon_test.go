package server

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/nazar256/intopwa/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tinyPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

type readSeekCloser struct {
	*bytes.Reader
}

func (r readSeekCloser) Close() error {
	return nil
}

func TestReadUploadedIconValidatesAndBuildsStableReference(t *testing.T) {
	header := &multipart.FileHeader{
		Filename: "client-name.png",
		Size:     int64(len(tinyPNG)),
		Header:   make(map[string][]string),
	}
	header.Header.Set("Content-Type", "image/png")

	icon, err := readUploadedIcon(readSeekCloser{Reader: bytes.NewReader(tinyPNG)}, header)
	require.NoError(t, err)

	assert.Equal(t, "image/png", icon.Props.MimeType)
	assert.Equal(t, "1x1", icon.Props.Size.String())
	assert.Equal(t, domain.UploadedIconsHost, icon.URL.Hostname())
	assert.Contains(t, icon.URL.Path, ".png")
	assert.NotContains(t, icon.URL.String(), "client-name")
	assert.Equal(t, tinyPNG, icon.Body)
}

func TestReadUploadedIconRejectsUnsupportedType(t *testing.T) {
	header := &multipart.FileHeader{
		Filename: "not-icon.txt",
		Size:     5,
		Header:   make(map[string][]string),
	}
	header.Header.Set("Content-Type", "text/plain")

	_, err := readUploadedIcon(readSeekCloser{Reader: bytes.NewReader([]byte("hello"))}, header)
	require.Error(t, err)
	assert.ErrorContains(t, err, errUploadedIconUnsupported.Error())
}

func TestReadUploadedIconRejectsOversizedBody(t *testing.T) {
	body := bytes.Repeat([]byte{0}, maxUploadedIconBytes+1)
	header := &multipart.FileHeader{
		Filename: "huge.png",
		Size:     int64(len(body)),
		Header:   make(map[string][]string),
	}
	header.Header.Set("Content-Type", "image/png")

	_, err := readUploadedIcon(readSeekCloser{Reader: bytes.NewReader(body)}, header)
	require.Error(t, err)
	assert.ErrorContains(t, err, errUploadedIconTooLarge.Error())
}

func TestNormalizedUploadedIconMimeTypeAcceptsDeclaredSVGWhenContentMatches(t *testing.T) {
	mimeType := normalizedUploadedIconMimeType([]byte(`<svg width="1" height="1"></svg>`), "image/svg+xml")
	assert.Equal(t, "image/svg+xml", mimeType)
}

func TestSupportedUploadedIconTypesIncludeBrowserIconTypes(t *testing.T) {
	for _, mimeType := range []string{"image/png", "image/jpeg", "image/gif", "image/webp", "image/svg+xml", "image/x-icon", "image/vnd.microsoft.icon"} {
		assert.True(t, isSupportedUploadedIconMimeType(mimeType), mimeType)
	}
	assert.False(t, isSupportedUploadedIconMimeType(http.DetectContentType([]byte("not image"))))
}
