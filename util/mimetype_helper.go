package util

import (
	"context"
	"mime"
	"strings"

	log "github.com/abeja-inc/abeja-platform-model-proxy/util/logging"
)

// DefaultExt is default extension for not binary file.
const DefaultExt = ".txt"

// DefaultBinaryExt is default extension for binary file.
const DefaultBinaryExt = ".bin"

var extMap = map[string]string{
	"application/javascript":            ".js",
	"application/json":                  ".json",
	"application/pdf":                   ".pdf",
	"application/x-www-form-urlencoded": ".txt",
	"audio/midi":                        ".midi",
	"audio/mpeg":                        ".mpg",
	"audio/ogg":                         ".oga",
	"audio/wav":                         ".wav",
	"audio/webm":                        ".webm",
	"image/bmp":                         ".bmp",
	"image/gif":                         ".gif",
	"image/jpeg":                        ".jpg",
	"image/png":                         ".png",
	"image/svg+xml":                     ".svg",
	"image/webp":                        ".webp",
	"text/csv":                          ".csv",
	"text/html":                         ".html",
	"text/markdown":                     ".md",
	"text/plain":                        ".txt",
	"text/xml":                          ".xml",
	"video/avi":                         ".avi",
	"video/mp4":                         ".mp4",
	"video/ogg":                         ".ogv",
	"video/quicktime":                   ".qt",
	"video/webm":                        ".webm",
	"video/x-matroska":                  ".mkv",
}

func GetExtension(ctx context.Context, contentType string) string {
	if contentType == "" {
		// GET request or form-data in multipart/form-data
		return DefaultExt
	}
	mt, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Warningf(ctx, "failed to parse `Content-Type` header: "+log.ErrorFormat, err)
		mt = ""
	}
	ext := extMap[mt]
	if ext == "" {
		if strings.HasPrefix(strings.ToLower(contentType), "text") {
			ext = DefaultExt
		} else {
			ext = DefaultBinaryExt
		}
	}
	return ext
}
