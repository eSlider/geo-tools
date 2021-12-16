package mbtiles

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test Tile
func TestTile(t *testing.T) {
	tile := &Tile{
		RowID:     0,
		ZoomLevel: 0,
		Column:    0,
		Row:       0,
		Data:      nil,
	}
	require.True(t, tile.IsEmpty())
}

func TestGetMeta(t *testing.T) {
	ex, err := NewExporter("../../data/tiles-world-vector.mbtiles", ExporterSettings{
		Path:       "tiles",
		Decompress: false,
	})
	require.NoError(t, err, "can't read mbtiles file")
	meta, err := ex.GetMeta()
	require.NoError(t, err, "can't read mbtiles file")
	require.NotEmptyf(t, meta.Name, "meta name is empty")
	require.NotEmpty(t, meta.Basename, "basename is empty")
}
