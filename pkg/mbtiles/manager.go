package mbtiles

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	db *sqlx.DB
}

func NewManager(path string) (*Manager, error) {
	var params = url.Values{}

	// Prevents any timeouts
	params.Add("_timeout", "0")

	// If shared-cache mode is enabled and a thread establishes multiple connections to the same database,
	// the connections share a single data and schema cache.
	// This can significantly reduce the quantity of memory and IO required by the system.
	params.Add("cache", "shared")

	// With synchronous OFF (0), SQLite continues without syncing as soon as
	// it has handed data off to the operating system.
	// If the application running SQLite crashes, the data will be safe,
	// but the database might become corrupted if the operating system crashes
	// or the computer loses power before that data has been written to the disk surface.
	// On the other hand, commits can be orders of magnitude faster with synchronous OFF.
	params.Add("_sync", "OFF")

	// No need to optimize database storage every time its changed
	params.Add("_vacuum", "0")
	params.Add("mode", "memory")
	params.Add("journal", "OFF")
	params.Add("immutable", "true")
	params.Add("_mutex", "full")
	params.Add("_query_only", "true")
	params.Add("_writable_schema", "false")
	params.Add("_fk", "false")
	params.Add("_defer_fk", "false")
	params.Add("_ignore_check_constraints", "true")
	params.Add("_cslike", "true")
	// params.Add("_cache_size", "0")
	dsn := path + "?" + params.Encode()
	db, err := sqlx.Open("sqlite3", dsn)

	if err != nil {
		return nil, err
	}
	m := Manager{
		db: db,
	}
	return &m, err
}

// GetTile data only a pbf image
func (m *Manager) GetTile(z int64, x int64, y int64) ([]byte, error) {
	var tileData []byte
	rows, err := m.db.Query(`
      SELECT "tile_data"
      FROM "tiles"
      WHERE "zoom_level"=?
        AND "tile_column"=?
        AND "tile_row"=?`, z, x, y)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan all result and append it
	for rows.Next() {
		var tmpTileData []byte
		if err := rows.Scan(&tmpTileData); err != nil {
			return nil, err
		}
		tileData = append(tileData, tmpTileData...)
	}

	return tileData, nil
}

// GetTiles list
func (m *Manager) GetTiles() ([]*Tile, error) {
	// Fetch tiles
	rows, err := m.db.Queryx("SELECT zoom_level, tile_column, tile_row FROM tiles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// sqlx.Next and sqlx.StructScan are not safe for concurrent use.
	// If you throw together a simple unit test for your code and run it
	// with the race detector go test -race, it will report a race condition
	// on an unexported field of the "database/sql".Rows struct:
	var tiles []*Tile
	for rows.Next() {
		var t Tile
		if err := rows.StructScan(&t); err != nil {
			logrus.WithError(err).
				Error("Fetch tile from db")
		}
		t.Data, err = m.GetTile(t.ZoomLevel, t.Column, t.Row)
		tiles = append(tiles, &t)
		if err != nil {
			logrus.WithField("tile", t).
				Warn("Get tile data")
		}
	}

	return tiles, nil
}

// WalkThroughTiles and call back by each tile
func (m *Manager) WalkThroughTiles(callback func(tile *Tile) bool, zoom int) error {
	// Fetch tiles
	rows, err := m.db.Queryx(fmt.Sprintf("SELECT zoom_level, tile_column, tile_row FROM tiles WHERE zoom_level = %d", zoom))
	if err != nil {
		return err
	}
	defer rows.Close()

	// sqlx.Next and sqlx.StructScan are not safe for concurrent use.
	// If you throw together a simple unit test for your code and run it
	// with the race detector go test -race, it will report a race condition
	// on an unexported field of the "database/sql".Rows struct:
	for rows.Next() {
		var t Tile
		if err := rows.StructScan(&t); err != nil {
			logrus.
				WithError(err).
				Error("Fetch tile from db")
		}
		t.Data, err = m.GetTile(t.ZoomLevel, t.Column, t.Row)
		if err != nil {
			logrus.
				WithField("tile", t).
				Warn("Get tile data")
			continue
		}
		continueWalk := callback(&t)
		t.Data = nil
		if !continueWalk {
			return nil
		}
	}
	return nil
}

// WalkThroughLayers and decode tile by the way
func (m *Manager) WalkThroughLayers(callback func(layer *mvt.Layer) bool, zoomLevel int) error {
	return m.WalkThroughTiles(func(tile *Tile) bool {
		pbf, err := tile.GetProtobuf()
		if err != nil {
			logrus.WithField("tile", tile).Warn("can't read tile proto buff")
			return true
		}
		layers, err := mvt.Unmarshal(pbf)
		if err != nil {
			logrus.WithField("tile", tile).Warn("can't unmarshal tile proto buff")
			return true
		}

		for _, layer := range layers {
			if !callback(layer) {
				return false
			}
		}
		return true
	}, zoomLevel)
}

// WalkThroughPlaces and call callback if something was found
func (m *Manager) WalkThroughPlaces(callback func(subj string, cls string, feature *geojson.Feature) bool) error {
	return m.WalkThroughLayers(func(layer *mvt.Layer) bool {
		if layer.Name != "place" {
			return true
		}
		for _, feature := range layer.Features {
			name, ok := feature.Properties["name:latin"].(string)
			if !ok {
				continue
			}
			cls, ok := feature.Properties["class"].(string)
			if !ok {
				continue
			}
			if !callback(name, cls, feature) {
				return false
			}
		}
		return true
	}, 14)
}

// Search features where subject has a query
func (m *Manager) Search(query string, maxResults int) ([]*geojson.Feature, error) {
	var features []*geojson.Feature
	return features, m.WalkThroughPlaces(func(subj string, cls string, feature *geojson.Feature) bool {
		if strings.Contains(strings.ToLower(subj), query) {
			features = append(features, feature)
		}
		return len(features) < maxResults
	})
}

// ExportGeoJson layer
func ExportGeoJson(layer *mvt.Layer) error {
	fc := ConvertMVTLayerToFeatureCollection(layer)
	gjData, err := json.Marshal(fc)
	if err != nil {
		return err
	}
	fcPath := fmt.Sprintf("tiles/%s.geo.json", layer.Name)
	if err = ioutil.WriteFile(fcPath, gjData, 0640); err != nil {
		return err
	}
	return nil
}

// ConvertMVTLayerToFeatureCollection to be exported as GeoJSON
func ConvertMVTLayerToFeatureCollection(layer *mvt.Layer) *geojson.FeatureCollection {
	return &geojson.FeatureCollection{
		Type: "FeatureCollection",
		ExtraMembers: map[string]interface{}{
			"version": layer.Version,
			"extent":  layer.Extent,
			"name":    layer.Name,
		},
		Features: layer.Features,
	}
}
