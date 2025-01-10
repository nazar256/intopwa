package icons

import (
	"encoding/json"
	"fmt"
	"github.com/nazar256/intopwa/internal/domain"
	"net/url"
	"slices"
	"strings"
)

//go:generate go run github.com/vektra/mockery/v2@v2.43.2 --dir=. --name kv --output ./mocks --outpkg mocks --case underscore  --with-expecter --exported
type kv interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
}

type cache struct {
	icons kv
}

func NewCache(icons kv) *cache {
	return &cache{
		icons: icons,
	}
}

func (c *cache) Get(urls []*url.URL) (icons []domain.Icon, found bool, err error) {
	// group icons by domain, domain is a key
	batches := make(map[string]struct{}, len(urls))
	for _, u := range urls {
		batches[u.Hostname()] = struct{}{}
	}

	var foundIcons []domain.Icon

	for host := range batches {
		iconsJSON, err := c.icons.Get(host)
		if err != nil {
			return icons, found, fmt.Errorf("failed to read from KV (key: %s): %w", host, err)
		}

		if iconsJSON == nil {
			continue
		}

		var iconsBatch []domain.Icon
		err = json.Unmarshal(iconsJSON, &iconsBatch)
		if err != nil {
			return icons, found, fmt.Errorf("failed to decode icons (%s): %w", host, err)
		}
		found = true

		foundIcons = append(foundIcons, iconsBatch...)
	}

	for _, i := range foundIcons {
		for _, u := range urls {
			if i.URL.String() == u.String() {
				icons = append(icons, i)
				break
			}
		}
	}

	return icons, found, nil
}

func (c *cache) Store(icons []domain.Icon) (err error) {
	// group icons by domain, domain is a key
	batches := make(map[string][]domain.Icon, 1)
	for _, icon := range icons {
		host := icon.URL.Hostname()
		batches[host] = append(batches[host], icon)
	}

	for host, iconsBatch := range batches {
		batchUrls := make([]*url.URL, 0, len(iconsBatch))
		for _, i := range iconsBatch {
			batchUrls = append(batchUrls, i.URL)
		}
		existingIcons, _, err := c.Get(batchUrls)
		if err != nil {
			return fmt.Errorf("failed to read icons cache before merge: %w", err)
		}

		existingIcons = append(existingIcons, iconsBatch...)

		slices.SortFunc(existingIcons, func(a, b domain.Icon) int {
			return strings.Compare(a.URL.String(), b.URL.String())
		})
		existingIcons = slices.CompactFunc(existingIcons, func(a, b domain.Icon) bool {
			return a.URL.String() == b.URL.String()
		})

		batches[host] = existingIcons
	}

	for host, batch := range batches {
		jsonValue, err := json.Marshal(batch)
		if err != nil {
			return fmt.Errorf("failed to encode icons batch: %w", err)
		}

		err = c.icons.Put(host, jsonValue)
		if err != nil {
			return fmt.Errorf("failed to write to KV: %w", err)
		}
	}

	return nil
}
