package links

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

//go:generate go run github.com/vektra/mockery/v2@v2.43.2 --dir=. --name kv --output ./mocks --outpkg mocks --case underscore  --with-expecter --exported
type kv interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
}

type cache struct {
	kv kv
}

func NewCache(kv kv) *cache {
	return &cache{
		kv: kv,
	}
}

func (c *cache) GetIconURLs(u *url.URL) (urls []*url.URL, found bool, err error) {
	key := keyFromURl(u)

	linksJSON, err := c.kv.Get(key)
	if err != nil {
		return urls, found, fmt.Errorf("failed to read from KV (key: %s): %w", key, err)
	}

	if linksJSON == nil {
		return urls, false, nil
	}

	var urlStrs []string
	err = json.Unmarshal(linksJSON, &urlStrs)
	if err != nil {
		return urls, found, fmt.Errorf("failed to decode kv (key:%s): %w", key, err)
	}

	for _, urlStr := range urlStrs {
		iconURL, err := url.Parse(urlStr)
		if err != nil {
			return urls, found, fmt.Errorf("failed to parse icon URL (%s): %w", urlStr, err)
		}
		urls = append(urls, iconURL)
	}

	return urls, true, nil
}

func (c *cache) StoreIconURLs(u *url.URL, iconsURLs []*url.URL) error {
	urls, _, err := c.GetIconURLs(u)
	if err != nil {
		return fmt.Errorf("failed to read icons list before merge: %w", err)
	}

	merged := make([]*url.URL, 0, len(iconsURLs)+len(urls))
	seen := make(map[string]struct{}, len(iconsURLs)+len(urls))

	for _, iu := range iconsURLs {
		if iu == nil {
			continue
		}

		urlStr := iu.String()
		if _, ok := seen[urlStr]; ok {
			continue
		}

		merged = append(merged, iu)
		seen[urlStr] = struct{}{}
	}

	for _, iu := range urls {
		if iu == nil {
			continue
		}

		urlStr := iu.String()
		if _, ok := seen[urlStr]; ok {
			continue
		}

		merged = append(merged, iu)
		seen[urlStr] = struct{}{}
	}

	urlStrs := make([]string, 0, len(merged))
	for _, iu := range merged {
		urlStrs = append(urlStrs, iu.String())
	}

	jsonValue, err := json.Marshal(urlStrs)
	if err != nil {
		return fmt.Errorf("failed to encode kv batch: %w", err)
	}

	err = c.kv.Put(keyFromURl(u), jsonValue)
	if err != nil {
		return fmt.Errorf("failed to write to KV (key:%s): %w", keyFromURl(u), err)
	}

	return nil
}

func keyFromURl(u *url.URL) string {
	key := u.Hostname() + strings.TrimSuffix(u.Path, "/")
	if u.RawQuery != "" {
		key += "?" + u.RawQuery
	}
	return key
}
