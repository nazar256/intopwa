package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/biessek/golang-ico"
	"github.com/gen2brain/svg"
	"github.com/nazar256/intopwa/internal/domain"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	_ "golang.org/x/image/webp"
)

const (
	uploadedIconFormField      = "iconFile"
	maxUploadedIconBytes       = 1024 * 1024
	maxUploadedIconRequestBody = maxUploadedIconBytes + 64*1024
)

var (
	errUploadedIconTooLarge     = errors.New("uploaded icon is too large")
	errUploadedIconUnsupported  = errors.New("uploaded icon type is unsupported")
	errUploadedIconInvalidImage = errors.New("uploaded icon is not a valid image")
)

type uploadedIconValidationError struct {
	message string
}

func (e uploadedIconValidationError) Error() string {
	return e.message
}

func readUploadedIcon(file multipart.File, header *multipart.FileHeader) (domain.Icon, error) {
	if file == nil || header == nil {
		return domain.Icon{}, uploadedIconValidationError{message: "No icon file was uploaded."}
	}

	if header.Size > maxUploadedIconBytes {
		return domain.Icon{}, uploadedIconValidationError{message: errUploadedIconTooLarge.Error()}
	}

	body, err := io.ReadAll(io.LimitReader(file, maxUploadedIconBytes+1))
	if err != nil {
		return domain.Icon{}, fmt.Errorf("failed to read uploaded icon: %w", err)
	}
	if len(body) > maxUploadedIconBytes {
		return domain.Icon{}, uploadedIconValidationError{message: errUploadedIconTooLarge.Error()}
	}

	props, err := uploadedIconProps(body, header.Header.Get("Content-Type"))
	if err != nil {
		if errors.Is(err, errUploadedIconTooLarge) || errors.Is(err, errUploadedIconUnsupported) || errors.Is(err, errUploadedIconInvalidImage) {
			return domain.Icon{}, uploadedIconValidationError{message: err.Error()}
		}
		return domain.Icon{}, err
	}

	iconURL, err := uploadedIconURL(body, props.MimeType)
	if err != nil {
		return domain.Icon{}, err
	}

	return domain.Icon{
		URL:   iconURL,
		Body:  body,
		Props: props,
	}, nil
}

func uploadedIconProps(body []byte, declaredContentType string) (domain.ImageProps, error) {
	if len(body) == 0 {
		return domain.ImageProps{}, errUploadedIconInvalidImage
	}

	mimeType := normalizedUploadedIconMimeType(body, declaredContentType)
	if !isSupportedUploadedIconMimeType(mimeType) {
		return domain.ImageProps{}, errUploadedIconUnsupported
	}

	cfg, err := decodeUploadedIconConfig(body, mimeType)
	if err != nil {
		return domain.ImageProps{}, fmt.Errorf("%w: %v", errUploadedIconInvalidImage, err)
	}

	if cfg.Width <= 0 || cfg.Height <= 0 {
		return domain.ImageProps{}, errUploadedIconInvalidImage
	}

	return domain.ImageProps{
		MimeType: mimeType,
		Size: domain.ImageSize{
			Width:  cfg.Width,
			Height: cfg.Height,
		},
	}, nil
}

func normalizedUploadedIconMimeType(body []byte, declaredContentType string) string {
	declaredContentType = strings.ToLower(strings.TrimSpace(strings.Split(declaredContentType, ";")[0]))
	if declaredContentType == "image/svg+xml" && looksLikeSVG(body) {
		return "image/svg+xml"
	}

	detected := strings.ToLower(http.DetectContentType(body))
	if detected == "text/xml; charset=utf-8" || detected == "text/plain; charset=utf-8" || detected == "application/octet-stream" {
		if looksLikeSVG(body) {
			return "image/svg+xml"
		}
	}

	return strings.Split(detected, ";")[0]
}

func looksLikeSVG(body []byte) bool {
	trimmed := strings.ToLower(string(bytes.TrimSpace(body)))
	return strings.Contains(trimmed, "<svg")
}

func isSupportedUploadedIconMimeType(mimeType string) bool {
	switch mimeType {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/x-icon", "image/vnd.microsoft.icon", "image/svg+xml":
		return true
	default:
		return false
	}
}

func decodeUploadedIconConfig(body []byte, mimeType string) (image.Config, error) {
	reader := bytes.NewReader(body)
	if mimeType == "image/x-icon" || mimeType == "image/vnd.microsoft.icon" {
		return ico.DecodeConfig(reader)
	}
	if mimeType == "image/svg+xml" {
		return svg.DecodeConfig(reader)
	}

	cfg, _, err := image.DecodeConfig(reader)
	return cfg, err
}

func uploadedIconURL(body []byte, mimeType string) (*url.URL, error) {
	hash := sha256.Sum256(body)
	path := "/" + hex.EncodeToString(hash[:]) + uploadedIconExtension(mimeType)
	return url.Parse("https://" + domain.UploadedIconsHost + path)
}

func uploadedIconExtension(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/x-icon", "image/vnd.microsoft.icon":
		return ".ico"
	default:
		return ".img"
	}
}
