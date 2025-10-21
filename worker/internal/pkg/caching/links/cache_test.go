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

func TestStoreIconURLsPrependsNewIcons(t *testing.T) {
	u, _ := url.Parse("https://chatgpt.com/")

	kv := newMemKV()
	existing := []string{
		"https://cdn-icons-png.flaticon.com/256/732/732242.png",
		"https://pngimg.com/d/chatgpt_PNG14.png",
	}

	existingJSON, err := json.Marshal(existing)
	require.NoError(t, err)
	err = kv.Put("chatgpt.com", existingJSON)
	require.NoError(t, err)

	cache := NewCache(kv)

	newIconURL, _ := url.Parse("https://static.chatgpt.com/icon.png")
	err = cache.StoreIconURLs(u, []*url.URL{newIconURL})
	require.NoError(t, err)

	urls, found, err := cache.GetIconURLs(u)
	require.NoError(t, err)
	require.True(t, found)

	if assert.Len(t, urls, 3) {
		assert.Equal(t, "https://static.chatgpt.com/icon.png", urls[0].String())
		assert.Equal(t, "https://cdn-icons-png.flaticon.com/256/732/732242.png", urls[1].String())
		assert.Equal(t, "https://pngimg.com/d/chatgpt_PNG14.png", urls[2].String())
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
