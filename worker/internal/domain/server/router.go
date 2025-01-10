package server

import (
	"context"
	"github.com/nazar256/intopwa/internal/domain"
	"net/http"
	"net/url"
	"strings"
)

//go:generate go run github.com/vektra/mockery/v2@v2.43.2 --name iconsFetcher --dir=. --output ./mocks --outpkg mocks --case underscore  --with-expecter --exported
type iconsFetcher interface {
	CacheIcons(ctx context.Context, pageURL *url.URL, iconURLs []*url.URL) error
	FetchIcons(ctx context.Context, u *url.URL) []domain.Icon
	One(ctx context.Context, iconURL *url.URL) (domain.Icon, error)
}

type server struct {
	iconsFetcher iconsFetcher
}

func New(iconsFetcher iconsFetcher) *server {
	return &server{
		iconsFetcher: iconsFetcher,
	}
}

func (s *server) Router() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Parse URL path and query
		urlPath := req.URL.Path
		switch {
		//case urlPath == "/favicon.ico":
		//	s.handleFavicon(w)
		case strings.HasPrefix(urlPath, "/a/"):
			s.handleApp(w, req)
		case strings.HasPrefix(urlPath, "/i/"):
			s.handleIcon(w, req)
		default:
			// If the path is not in the expected format, return a default response
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}
	}
}
