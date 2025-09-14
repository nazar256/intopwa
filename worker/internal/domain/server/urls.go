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
	path := u.appPath() + redirectPagePath
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	return path
}

func (u *appURL) manifestPath() string {
	path := u.appPath() + manifestPath
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	return path
}

func (u *appURL) serviceWorkerPath() string {
	path := u.appPath() + serviceWorkerPath
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	return path
}
