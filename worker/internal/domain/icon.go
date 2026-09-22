package domain

import (
	"net/url"
	"strconv"
	"strings"
)

// UploadedIconsHost is the synthetic host used for icon URLs that reference
// uploaded icon files stored in the worker's KV icon cache. These URLs never
// resolve over the network; they are served by the worker's /i/ endpoint from
// cache only.
const UploadedIconsHost = "uploaded-icons.intopwa.local"

type Icon struct {
	URL   *url.URL
	Body  []byte
	Props ImageProps
}

// Name returns the icon filename without path or query string
func (i Icon) Name() string {
	return i.URL.Path
}
func (i Icon) Path() string {
	pathParts := []string{"/i/", i.URL.Hostname()}
	if i.URL.Port() != "" {
		pathParts = append(pathParts, ":", i.URL.Port())
	}

	pathParts = append(pathParts, i.URL.Path)
	if i.URL.RawQuery != "" {
		pathParts = append(pathParts, "?", i.URL.RawQuery)
	}

	return strings.Join(pathParts, "")
}

type ImageProps struct {
	MimeType string
	Size     ImageSize
}

type ImageSize struct {
	Width  int
	Height int
}

func (s ImageSize) String() string {
	return strconv.Itoa(s.Width) + "x" + strconv.Itoa(s.Height)
}
