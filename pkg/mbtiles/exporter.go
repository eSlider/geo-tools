package mbtiles

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"

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
	Manager

	cfg ExporterSettings
	wg  sync.WaitGroup

	tiles      []*Tile
	TilesCount int
}

// NewExporter creates a new sitemap exporter.
func NewExporter(importPath string, settings ExporterSettings) (*Exporter, error) {
	manager, err := NewManager(importPath)
	if err != nil {
		return nil, err
	}
	return &Exporter{
		Manager: *manager,
		cfg:     settings,
	}, nil
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
