package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

func (s *server) handleIcon(w http.ResponseWriter, req *http.Request) {
	iconU, err := parseIconURL(req.URL)
	if err != nil {
		slog.Error("failed to parse icon URL", "err", err)
		// If the path is not in the expected format, return a default response
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	icon, err := s.iconsFetcher.One(req.Context(), iconU)
	if err != nil {
		slog.Error("failed to fetch icon", "err", err, "URL", iconU.String())
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", icon.Props.MimeType)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(icon.Body)
	if err != nil {
		slog.Error("failed to write icon response", "err", err, "iconURL", iconU.String())
	}
}

func parseIconURL(u *url.URL) (*url.URL, error) {
	if u == nil {
		return nil, errors.New("URL is nil")
	}

	if !strings.HasPrefix(u.Path, "/i/") {
		return nil, errors.New("Invalid request format")
	}

	// Remove the "/i/" prefix from the path
	iconURLstr := strings.TrimPrefix(u.Path, "/i/")
	if u.RawQuery != "" {
		iconURLstr += "?" + u.RawQuery
	}

	// Add https:// scheme and parse the URL
	fullURL, err := url.Parse("https://" + iconURLstr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse icon URL: %w", err)
	}

	return fullURL, nil
}
