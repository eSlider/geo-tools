package mbtiles

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

type Place struct {
	ID         interface{}
	Type       string
	Geometry   *orb.Geometry
	Properties *geojson.Properties

	Class     string `json:"class" mapstructure:"class"`
	NameLatin string `json:"name:latin" mapstructure:"name:latin"`
}
