package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type pwaManifest struct {
	Name            string    `json:"name,omitempty"`
	ShortName       string    `json:"short_name,omitempty"`
	Icons           []pwaIcon `json:"icons,omitempty"`
	StartURL        string    `json:"start_url"`
	BackgroundColor string    `json:"background_color,omitempty"`
	ThemeColor      string    `json:"theme_color,omitempty"`
	Display         string    `json:"display"`
}

type pwaIcon struct {
	Src   string `json:"src"`
	Type  string `json:"type"`
	Sizes string `json:"sizes"`
}

func (s *server) handleAppRoot(ctx context.Context, w http.ResponseWriter, u *appURL, iconURLs []*url.URL) {
	err := s.iconsFetcher.CacheIcons(ctx, &u.URL, iconURLs)
	if err != nil {
		slog.Error("failed to cache icons", "err", err)
	}

	w.Header().Set("Content-Type", "text/html")

	infoPage := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">

		<meta name="theme-color" content="#317EFB"/>
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>%s</title>
		<link rel="manifest" href="%s">
		<link rel="apple-touch-icon" href="icons/icon-192x192.png">
		<link rel="icon" type="image/x-icon" href="/favicon.ico">
		<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
		<link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png">
    	<link rel="stylesheet" href="/styles.css">
		<script>
			if ('serviceWorker' in navigator) {
				navigator.serviceWorker.register('%s')
					.then(function(registration) {
						console.log('Service Worker registered with scope:', registration.scope);
					}).catch(function(error) {
					console.log('Service Worker registration failed:', error);
				});
			}
		</script>
	</head>
	<body>
		<div class="container">
			<h2>Install this app</h2>
			<h3>Mobile:</h3>
			<p>Tap the browser menu (⋮) and select 'Add to Home Screen' or 'Install App'</p>
			<h3>Desktop:</h3>
			<p>Click the install icon (⇩) in your browser's address bar</p>
		</div>
	</body>
	</html>
`, "App for "+u.URL.Hostname(), u.manifestPath(), u.serviceWorkerPath())

	_, err = fmt.Fprintln(w, infoPage)
	if err != nil {
		slog.Error("failed to write app root", "err", err)
	}
}

func (s *server) handleManifest(ctx context.Context, w http.ResponseWriter, appURL *appURL) {
	title := fmt.Sprintf(appURL.URL.Hostname() + appURL.URL.Path)

	icons := s.iconsFetcher.FetchIcons(ctx, &appURL.URL)

	var pwaIcons []pwaIcon
	for _, icon := range icons {
		pwaIcons = append(pwaIcons, pwaIcon{
			Src:   icon.Path(),
			Type:  icon.Props.MimeType,
			Sizes: icon.Props.Size.String(),
		})
	}
	pwaIcons = ensureAnyIcon(pwaIcons)

	manifest := pwaManifest{
		Name:            title,
		ShortName:       title,
		Icons:           pwaIcons,
		StartURL:        appURL.redirectPagePath(),
		BackgroundColor: "#3367D6",
		ThemeColor:      "#3367D6",
		Display:         "standalone",
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(manifest)
	if err != nil {
		slog.Error("failed to write manifest response", "err", err)
	}
}

func (s *server) handleServiceWorker(w http.ResponseWriter, u *appURL) {
	swScript := `self.addEventListener('install', event => {
		event.waitUntil(
			caches.open('v1').then(cache => {
				return cache.addAll(['` + u.redirectPagePath() + `']);
			})
		);
	});
	self.addEventListener('fetch', event => {
		event.respondWith(
			caches.match(event.request).then(response => {
				return response || fetch(event.request);
			})
		);
	});`

	w.Header().Set("Content-Type", "application/javascript")
	_, err := fmt.Fprintln(w, swScript)
	if err != nil {
		slog.Error("failed to write service worker script", "err", err)
	}
}

func (s *server) handleRedirect(w http.ResponseWriter, u *appURL) {
	redirectHTML := fmt.Sprintf(`<html><head><meta http-equiv="refresh" content="0;url=%s"></head></html>`, u.String())

	w.Header().Set("Content-Type", "text/html")
	_, err := fmt.Fprintln(w, redirectHTML)
	if err != nil {
		slog.Error("failed to write redirect page", "err", err)
	}
}

func ensureAnyIcon(icons []pwaIcon) []pwaIcon {
	if len(icons) > 0 {
		return icons
	}

	return append(icons, pwaIcon{
		Src:   "/default-app-icon.png",
		Type:  "image/png",
		Sizes: "512x512",
	})
}
