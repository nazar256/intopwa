package server

import (
	"errors"
	"fmt"
	"github.com/nazar256/intopwa/internal/domain"
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
		var uploadedIcon *domainIconUpload
		if req.Method == http.MethodPost {
			iconURLs, uploadedIcon, err = s.parseCreateAppIconSources(w, req)
			if err != nil {
				if validationErr, ok := err.(uploadedIconValidationError); ok {
					http.Error(w, validationErr.Error(), http.StatusBadRequest)
					return
				}

				slog.Error("failed to parse icon sources", "err", err)
				http.Error(w, "Invalid icon upload", http.StatusBadRequest)
				return
			}

			if uploadedIcon != nil {
				err = s.iconsFetcher.StoreUploadedIcon(ctx, &appU.URL, uploadedIcon.icon)
				if err != nil {
					slog.Error("failed to store uploaded icon", "err", err)
					http.Error(w, "Failed to store uploaded icon", http.StatusInternalServerError)
					return
				}
			}
		}
		s.handleAppRoot(ctx, w, appU, iconURLs)
	}
}

type domainIconUpload struct {
	icon domain.Icon
}

func (s *server) parseCreateAppIconSources(w http.ResponseWriter, req *http.Request) ([]*url.URL, *domainIconUpload, error) {
	if strings.HasPrefix(req.Header.Get("Content-Type"), "multipart/form-data") {
		req.Body = http.MaxBytesReader(w, req.Body, maxUploadedIconRequestBody)
		if err := req.ParseMultipartForm(maxUploadedIconRequestBody); err != nil {
			return nil, nil, uploadedIconValidationError{message: "Uploaded icon request is too large or malformed."}
		}

		iconURLs, err := parseIconURLValues(req.MultipartForm.Value["icons[]"])
		if err != nil {
			return nil, nil, err
		}

		file, header, err := req.FormFile(uploadedIconFormField)
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				return iconURLs, nil, nil
			}
			return nil, nil, err
		}
		defer file.Close()

		if len(iconURLs) > 0 {
			return nil, nil, uploadedIconValidationError{message: "Choose either icon URLs or one uploaded icon file, not both."}
		}

		icon, err := readUploadedIcon(file, header)
		if err != nil {
			return nil, nil, err
		}

		return nil, &domainIconUpload{icon: icon}, nil
	}

	if err := req.ParseForm(); err != nil {
		return nil, nil, err
	}

	iconURLs, err := parseIconURLValues(req.Form["icons[]"])
	return iconURLs, nil, err
}

func parseIconURLValues(values []string) ([]*url.URL, error) {
	iconURLs := make([]*url.URL, 0, len(values))
	for _, iconURLStr := range values {
		iconURLStr = strings.TrimSpace(iconURLStr)
		if iconURLStr == "" {
			continue
		}
		if !strings.HasPrefix(iconURLStr, "http://") && !strings.HasPrefix(iconURLStr, "https://") {
			iconURLStr = "https://" + iconURLStr
		}
		iconURL, err := url.Parse(iconURLStr)
		if err != nil {
			return nil, uploadedIconValidationError{message: "Invalid icon URL."}
		}
		iconURLs = append(iconURLs, iconURL)
	}
	return iconURLs, nil
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

	query := u.Query()
	if strings.HasSuffix(urlPath, manifestPath) {
		query.Del("v")
	}

	if encoded := query.Encode(); encoded != "" {
		base += "?" + encoded
	}

	parsedURL, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("failed to parse app URL: %w", err)
	}

	return &appURL{URL: *parsedURL}, nil
}
