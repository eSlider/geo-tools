package tiles

import (
	"errors"
	"fmt"
	"math"
)

// ErrEmptyTileData error
var ErrEmptyTileData = errors.New("tile is empty")

// Tile represents a tile from MB Tiles file
type Tile struct {
	RowID     int    `db:"ROWID"`
	ZoomLevel int64  `db:"zoom_level" geo:"z"`
	Column    int64  `db:"tile_column" geo:"x"`
	Row       int64  `db:"tile_row" geo:"y"`
	Data      []byte `db:"tile_data"`

	// Internal Number
	Number int
}

// IsEmpty data?
func (t *Tile) IsEmpty() bool {
	return len(t.Data) > 0
}

// GetPath of tile
func (t *Tile) GetPath() string {
	return fmt.Sprintf("%d/%d", t.ZoomLevel, t.Column)
}

// GetFileName XYZtoEPSG
func (t *Tile) GetFileName() int64 {
	return int64(math.Pow(2, float64(t.ZoomLevel)) - 1 - float64(t.Row))
}

// GetFormat of a tile
func (t *Tile) GetFormat() (TileFormat, error) {
	if t.IsEmpty() {
		return UNKNOWN, ErrEmptyTileData
	}
	return DetectTileFormat(t.Data)
}

// DetectTileFormat by data prefix
func (t *Tile) DetectTileFormat() (TileFormat, error) {
	return DetectTileFormat(t.Data)
}
