package tiles

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test Tile
func TestFetchToplist2(t *testing.T) {
	tile := &Tile{
		RowID:     0,
		ZoomLevel: 0,
		Column:    0,
		Row:       0,
		Data:      nil,
	}
	require.True(t, tile.IsEmpty())
}
