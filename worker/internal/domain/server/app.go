package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

const (
	manifestPath      = "/manifest.json"
	serviceWorkerPath = "/service-worker.js"
	redirectPagePath  = "/redirect.html"
)

func (s *server) handleApp(w http.ResponseWriter, req *http.Request) {
	urlPath := req.URL.Path

	parts := strings.SplitN(urlPath, "/", 3)
	if len(parts) < 3 {
		//If the path is not in the expected format, return a default response
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	appU, err := parseAppURL(req.URL)
	if err != nil {
		slog.Error("failed to parse app URL", "err", err)
		// If the path is not in the expected format, return a default response
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	switch {
	case strings.HasSuffix(urlPath, manifestPath):
		s.handleManifest(ctx, w, appU)
	case strings.HasSuffix(urlPath, serviceWorkerPath):
		s.handleServiceWorker(w, appU)
	case strings.HasSuffix(urlPath, redirectPagePath):
		s.handleRedirect(w, appU)
	default:
		// Default handler: show info page with links to manifest and service worker

		var iconURLs []*url.URL
		if req.Method == http.MethodPost {
			err = req.ParseForm()
			if err != nil {
				slog.Error("failed to parse form", "err", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			for _, iconURLStr := range req.Form["icons[]"] {
				if !strings.HasPrefix(iconURLStr, "http://") && !strings.HasPrefix(iconURLStr, "https://") {
					iconURLStr = "https://" + iconURLStr
				}
				iconURL, err := url.Parse(iconURLStr)
				if err != nil {
					slog.Error("failed to parse icon URL", "err", err)
					continue
				}
				iconURLs = append(iconURLs, iconURL)
			}
		}
		s.handleAppRoot(ctx, w, appU, iconURLs)
	}
}
func parseAppURL(u *url.URL) (*appURL, error) {
	urlPath := u.Path

	parts := strings.SplitN(urlPath, "/", 3)
	if len(parts) < 3 {
		return nil, errors.New("Invalid request format")
	}

	appURLValue := parts[2]

	fileSuffixes := []string{
		manifestPath,
		serviceWorkerPath,
		redirectPagePath,
	}

	for _, suffix := range fileSuffixes {
		appURLValue = strings.TrimSuffix(appURLValue, suffix)
	}

	base := "https://" + appURLValue
	if u.RawQuery != "" {
		base += "?" + u.RawQuery
	}

	parsedURL, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("failed to parse app URL: %w", err)
	}

	return &appURL{URL: *parsedURL}, nil
}
