package server

import (
	"net/url"
	"strings"
)

type appURL struct {
	url.URL
}

func (u *appURL) appPath() string {
	pathParts := []string{"/a/", u.Hostname()}
	if u.Port() != "" {
		pathParts = append(pathParts, ":", u.Port())
	}

	pathParts = append(pathParts, strings.TrimSuffix(u.Path, "/"))

	return strings.Join(pathParts, "")
}

func (u *appURL) redirectPagePath() string {
	return u.appPath() + redirectPagePath
}

func (u *appURL) manifestPath() string {
	return u.appPath() + manifestPath
}

func (u *appURL) serviceWorkerPath() string {
	return u.appPath() + serviceWorkerPath
}
