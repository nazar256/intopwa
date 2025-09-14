package links

import (
	"net/url"
	"testing"

	"github.com/nazar256/intopwa/internal/pkg/caching/links/mocks"
	"github.com/stretchr/testify/assert"
)

func TestKeyFromURLIncludesQuery(t *testing.T) {
	u1, _ := url.Parse("https://example.com/path?foo=bar")
	u2, _ := url.Parse("https://example.com/path?foo=baz")
	assert.NotEqual(t, keyFromURl(u1), keyFromURl(u2))
}

func TestGetIconURLsUsesQueryInKey(t *testing.T) {
	u, _ := url.Parse("https://example.com/path?foo=bar")
	key := "example.com/path?foo=bar"

	kv := mocks.NewKv(t)
	kv.EXPECT().Get(key).Return([]byte(`["https://icon.com/favicon.ico"]`), nil).Once()

	c := NewCache(kv)
	urls, found, err := c.GetIconURLs(u)
	assert.NoError(t, err)
	assert.True(t, found)
	if assert.Len(t, urls, 1) {
		assert.Equal(t, "https://icon.com/favicon.ico", urls[0].String())
	}
}
