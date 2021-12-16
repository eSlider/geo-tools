package mbtiles

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mitchellh/mapstructure"
	_ "github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

// ExporterSettings groups all cfg for NewExporter
type ExporterSettings struct {
	Path       string
	Decompress bool
	BaseUrl    string
}

// Exporter of mbtiles
type Exporter struct {
	db  *sqlx.DB
	cfg ExporterSettings
	wg  sync.WaitGroup

	tiles      []*Tile
	TilesCount int
}

// NewExporter creates a new sitemap exporter.
func NewExporter(importPath string, settings ExporterSettings) (*Exporter, error) {
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

	// No need to optimize database storage every time it's changes
	params.Add("_vacuum", "0")
	params.Add("mode", "memory")
	params.Add("journal", "OFF")
	params.Add("immutable", "true")
	params.Add("_mutex", "no")
	params.Add("_query_only", "true")
	params.Add("_writable_schema", "false")
	params.Add("_fk", "false")
	params.Add("_defer_fk", "false")
	params.Add("_ignore_check_constraints", "true")
	dsn := importPath + "?" + params.Encode()
	db, err := sqlx.Open("sqlite3", dsn)

	if err != nil {
		return nil, err
	}

	return &Exporter{
		db:  db,
		cfg: settings,
	}, nil
}

// GetTile data only a pbf image
func (ex *Exporter) GetTile(z int64, x int64, y int64) ([]byte, error) {
	var tileData []byte
	rows, err := ex.db.Query(`
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

// Export tiles
func (ex *Exporter) Export() error {
	var err error

	// Fetch tiles
	if ex.tiles, err = ex.GetTiles(); err != nil {
		return err
	}

	ex.TilesCount = len(ex.tiles)
	ex.wg.Add(ex.TilesCount)
	for id, tile := range ex.tiles {
		tile.Number = id + 1
		go ex.exportTile(tile)
	}
	ex.wg.Wait()

	// Get meta
	meta, err := ex.GetMeta()
	if err != nil {
		return err
	}

	// Generate JSON from meta
	indexJson, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	// Write index file
	indexPath := fmt.Sprintf("%s/index.json", ex.cfg.Path)
	if err := os.WriteFile(indexPath, indexJson, 0600); err != nil {
		logrus.
			WithField("path", indexPath).
			WithError(err).
			Error("Write index file")
	}
	logrus.
		WithField("path", indexPath).
		Info("Write index file")

	return nil
}

// GetTiles list
func (ex *Exporter) GetTiles() ([]*Tile, error) {
	// Fetch tiles
	rows, err := ex.db.Queryx("SELECT zoom_level, tile_column, tile_row FROM tiles")
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
			logrus.
				WithError(err).
				Error("Fetch tile from db")
		}
		t.Data, err = ex.GetTile(t.ZoomLevel, t.Column, t.Row)
		tiles = append(tiles, &t)
		if err != nil {
			logrus.
				WithField("tile", t).
				Warn("Get tile data")
		}
	}

	return tiles, nil
}

// exportTile
func (ex *Exporter) exportTile(t *Tile) {
	defer ex.wg.Done()

	// Create tilesPath
	tilesPath := fmt.Sprintf("%s/%s", ex.cfg.Path, t.GetPath())
	if err := os.MkdirAll(tilesPath, 0750); err != nil {
		logrus.
			WithField("path", tilesPath).
			WithError(err)
	}
	// Get tile data
	// tileData, err := ex.GetTile(t.ZoomLevel, t.Column, t.Row)
	// if err != nil {
	// 	logrus.
	// 		WithField("tile", t).
	// 		Warn("Get tile data")
	// }

	// Detect tile data
	var tileDataReader io.Reader
	tileDataReader = bytes.NewReader(t.Data)
	fileType := "pbf"
	format, err := t.DetectTileFormat()
	if err != nil {
		logrus.
			WithField("tile", t).
			Warn("Detect tile format")
	}

	// Decompress depending on the format
	switch format {
	case GZIP:
		if ex.cfg.Decompress {
			tileDataReader, _ = gzip.NewReader(tileDataReader)
		}
	case ZLIB:
		if ex.cfg.Decompress {
			tileDataReader, _ = zlib.NewReader(tileDataReader)
		}
	case JPG:
		fileType = "jpg"
	case PNG:
		fileType = "png"
	case WEBP:
		fileType = "webp"
	}

	// Defile tile file name
	tileFileName := fmt.Sprintf("%s/%d.%s", tilesPath, t.GetFileName(), fileType)

	// Read tile data
	pbf, err := ioutil.ReadAll(tileDataReader)
	if err != nil {
		logrus.
			WithField("path", tileFileName).
			WithError(err).Error("Read tile data")
	}

	// Write tile file
	if err := os.WriteFile(tileFileName, pbf, 0600); err != nil {
		logrus.
			WithField("path", tileFileName).
			WithError(err).Error("Write tile file")
	}

	// Clean memory from PBF
	t.Data = nil

	// Inform
	logrus.
		WithField("from", t.Number).
		WithField("count", ex.TilesCount).
		WithField("path", tileFileName).
		Info("Extract tile file")
}

// GetMeta data from database file
func (ex *Exporter) GetMeta() (*Meta, error) {
	rows, err := ex.db.Queryx("SELECT name,value FROM metadata")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metaMap := map[string]string{}
	for rows.Next() {
		var k, v string
		err = rows.Scan(&k, &v)
		metaMap[k] = v
	}

	meta := &Meta{
		Scheme:   "xyz",
		Type:     "baselayer",
		TileJson: "2.0.0",
		Format:   "pbf",
		Basename: "base",
		Profile:  "mercator",
		Scale:    1,
		Tiles:    []string{fmt.Sprintf("%s/{z}/{x}/{y}.pbf", ex.cfg.BaseUrl)},
		Bounds:   stringToFloatArray(metaMap["bounds"]),
		Center:   stringToFloatArray(metaMap["center"]),
	}

	if err := json.Unmarshal([]byte(metaMap["json"]), meta); err != nil {
		return nil, err
	}

	if err := mapstructure.WeakDecode(&metaMap, meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func stringToFloatArray(floats string) []float64 {
	split := strings.Split(floats, ",")
	var bounds = make([]float64, len(split))
	for i, bound := range split {
		bf, _ := strconv.ParseFloat(bound, 32)
		bounds[i] = bf
	}
	return bounds
}
