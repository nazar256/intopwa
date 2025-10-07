package server

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestAppURLPathsIncludeQueryAndPort(t *testing.T) {
	u := &appURL{URL: url.URL{Scheme: "https", Host: "www.windy.com:8443", Path: "/test/meteogram", RawQuery: "49"}}

	assert.Equal(t, "/a/www.windy.com:8443/test/meteogram/manifest.json?49", u.manifestPath())
	assert.Equal(t, "/a/www.windy.com:8443/test/meteogram/service-worker.js?49", u.serviceWorkerPath())
	assert.Equal(t, "/a/www.windy.com:8443/test/meteogram/redirect.html?49", u.redirectPagePath())
}
