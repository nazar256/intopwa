//go:build cloudflare
// +build cloudflare

package cloudflare

import (
	"fmt"
	"github.com/syumai/workers/cloudflare/fetch"
	"net/http"
)

type fetcher struct {
	client *fetch.Client
}

func NewFetcher() *fetcher {
	return &fetcher{
		client: fetch.NewClient(),
	}
}

func (f *fetcher) Do(req *http.Request) (resp *http.Response, err error) {
	cfReq, err := fetch.NewRequest(
		req.Context(),
		req.Method,
		req.URL.String(),
		req.Body,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	resp, err = f.client.Do(cfReq, &fetch.RequestInit{
		CF:       &fetch.RequestInitCF{},
		Redirect: fetch.RedirectModeFollow,
	})
	if err != nil {
		return resp, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	return resp, err
}
