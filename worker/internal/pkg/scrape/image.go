package scrape

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gen2brain/svg"
	"github.com/nazar256/intopwa/internal/domain"
	"image"
	"net/http"
	"strings"

	"github.com/biessek/golang-ico"
	_ "golang.org/x/image/webp"
	_ "image/png"
)

func decodeImgProps(img []byte, mimeType string) (domain.ImageProps, error) {
	if len(img) == 0 {
		return domain.ImageProps{}, errors.New("empty image data passed")
	}

	if mimeType == "" {
		mimeType = http.DetectContentType(img)
	}
	props := domain.ImageProps{
		MimeType: mimeType,
	}
	var cfg image.Config
	var err error
	// Decode the image config to get the size
	if strings.Contains(mimeType, "x-icon") {
		// Use the ICO decoder for ICO images
		cfg, err = ico.DecodeConfig(bytes.NewReader(img))
	} else if strings.Contains(mimeType, "/svg") {
		cfg, err = svg.DecodeConfig(bytes.NewReader(img))
	} else {
		// Use the standard library's decoder for other image formats
		cfg, _, err = image.DecodeConfig(bytes.NewReader(img))
	}
	if err != nil {
		return props, fmt.Errorf("failed to decode image config: %w", err)
	}

	props.Size = domain.ImageSize{
		Width:  cfg.Width,
		Height: cfg.Height,
	}

	return props, nil
}
