package links

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/nazar256/intopwa/internal/pkg/caching/links/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestStoreIconURLsFiltersNilAndDuplicates(t *testing.T) {
	u, _ := url.Parse("https://chatgpt.com/")

	kv := newMemKV()
	cache := NewCache(kv)

	icon1, _ := url.Parse("https://static.chatgpt.com/icon.png")
	icon2, _ := url.Parse("https://static.chatgpt.com/icon.svg")

	err := cache.StoreIconURLs(u, []*url.URL{icon1, nil, icon1, icon2})
	require.NoError(t, err)

	urls, found, err := cache.GetIconURLs(u)
	require.NoError(t, err)
	require.True(t, found)

	if assert.Len(t, urls, 2) {
		assert.Equal(t, "https://static.chatgpt.com/icon.png", urls[0].String())
		assert.Equal(t, "https://static.chatgpt.com/icon.svg", urls[1].String())
	}
}

func TestStoreIconURLsReplacesExistingIcons(t *testing.T) {
	u, _ := url.Parse("https://chatgpt.com/")

	kv := newMemKV()
	existing := []string{
		"https://default.cdn.com/icon.png",
	}

	existingJSON, err := json.Marshal(existing)
	require.NoError(t, err)
	err = kv.Put("chatgpt.com", existingJSON)
	require.NoError(t, err)

	cache := NewCache(kv)

	customIcon, _ := url.Parse("https://static.chatgpt.com/custom.png")
	err = cache.StoreIconURLs(u, []*url.URL{customIcon})
	require.NoError(t, err)

	urls, found, err := cache.GetIconURLs(u)
	require.NoError(t, err)
	require.True(t, found)

	if assert.Len(t, urls, 1) {
		assert.Equal(t, "https://static.chatgpt.com/custom.png", urls[0].String())
	}
}

type memKV struct {
	data map[string][]byte
}

func newMemKV() *memKV {
	return &memKV{data: make(map[string][]byte)}
}

func (m *memKV) Get(key string) ([]byte, error) {
	value, ok := m.data[key]
	if !ok {
		return nil, nil
	}
	return value, nil
}

func (m *memKV) Put(key string, value []byte) error {
	m.data[key] = value
	return nil
}
