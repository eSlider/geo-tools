package spatialite

import (
	"database/sql"
	"testing"

	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// Test Tile
func TestSpatialite(t *testing.T) {
	sql.Register("sqlite3_with_spatialite",
		&sqlite3.SQLiteDriver{
			Extensions: []string{"mod_spatialite"},
		})
	spatialite, err := sql.Open("sqlite3_with_spatialite", ":memory:")
	require.NoError(t, err, "can't load spatialite driver")
	defer func(spatialite *sql.DB) {
		if err := spatialite.Close(); err != nil {
			require.NoError(t, err, "can't close spatialite")
		}
	}(spatialite)
	for _, query := range []string{
		"SELECT InitSpatialMetaData(1);",
		"DROP TABLE IF EXISTS testtable",
		"CREATE TABLE testtable (id INTEGER PRIMARY KEY AUTOINCREMENT, name CHAR(255));",
		"SELECT AddGeometryColumn('testtable', 'geom', 4326, 'POLYGON', 2);",
		"SELECT CreateSpatialIndex('testtable', 'geom');",
		"INSERT INTO testtable (name, geom) VALUES ('Test', GeomFromText('POLYGON((10 10, 20 10, 20 20, 10 20, 10 10))', 4326));",
	} {
		_, err = spatialite.Exec(query)
		require.NoError(t, err, err.Error())
	}
}
