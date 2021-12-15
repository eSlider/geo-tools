package tiles

import (
	"bytes"
	"errors"
)

// TileFormat of a tile
type TileFormat int

// List of possible formats
const (
	UNKNOWN = iota
	GZIP
	ZLIB
	PNG
	JPG
	WEBP
)

// ErrUnknownTileFormat an error of unknown tile format
var ErrUnknownTileFormat = errors.New("could not detect tile format")

// TypeFormatPatterns of any possible tile
var TypeFormatPatterns = map[TileFormat][]byte{
	GZIP: []byte("\x1f\x8b"), // this masks PBF format too
	ZLIB: []byte("\x78\x9c"),
	PNG:  []byte("\x89\x50\x4E\x47\x0D\x0A\x1A\x0A"),
	JPG:  []byte("\xFF\xD8\xFF"),
	WEBP: []byte("\x52\x49\x46\x46"),
}

// DetectTileFormat by data slice
func DetectTileFormat(data []byte) (TileFormat, error) {
	for format, pattern := range TypeFormatPatterns {
		if bytes.HasPrefix(data, pattern) {
			return format, nil
		}
	}

	return UNKNOWN, ErrUnknownTileFormat
}
