package mbtiles

// Meta of data.
// The metadata table MAY contain additional rows for tile sets that implement UTFGrid-based interaction or for other purposes.
// see: https://github.com/mapbox/mbtiles-spec/blob/master/1.3/spec.md
type Meta struct {
	// The human-readable name of the tile set.
	Name string `json:"name,omitempty"`

	// Tile set big description
	Description string `json:"description,omitempty"`

	// Tile set file name or relative path to the file.
	Basename string `json:"basename,omitempty"`

	// Tile set version.
	Version string `json:"version,omitempty"`

	// The file format of the tile data: `pbf`, `jpg`, `png`, `webp`, or an IETF media type for other formats.
	// `pbf` as a format refers to gzip-compressed vector tile data in Mapbox Vector Tile format.
	Format string `json:"format,omitempty"`

	// ???
	Type string `json:"type,omitempty"`

	// ???
	Scale int `json:"scale,omitempty"`

	// The lowest zoom level for which the tile set provides data
	MinZoom int `json:"minzoom,omitempty"`

	// The highest zoom level for which the tileset provides data
	MaxZoom int `json:"maxzoom,omitempty"`

	// The longitude, latitude, and zoom level of the default view of the map.
	// Example: -122.1906,37.7599,11
	Center []float64 `json:"center,omitempty" mapstructure:"-"`

	// The maximum extent of the rendered map area.
	// Bounds must define an area covered by all zoom levels.
	// The bounds are represented as WGS 84 latitude and longitude values, in the OpenLayers Bounds format (left, bottom, right, top).
	// For example, the bounds of the full Earth, minus the poles, would be: -180.0,-85,180,85.
	Bounds []float64 `json:"bounds,omitempty" mapstructure:"-"`

	// The `JSON` object in the json row MUST contain a `vector_layers` key, whose value is an array of JSON objects.
	// Each of those JSON objects describes one layer of vector tile data, and MUST contain the following key-value pairs:
	VectorLayers []VectorLayer `json:"vector_layers,omitempty"`

	// ???
	Profile string `json:"profile,omitempty"`

	// PBF's URL template. Example: ["http://localhost/tiles/{z}/{x}/{y}.pbf"]
	Tiles []string `json:"tiles,omitempty"`

	// ???
	TileJson string `json:"tile-json,omitempty"`

	// Example: "xyz"
	Scheme string `json:"scheme,omitempty"`

	// Database raw json contains vector layer info
	// The JSON object in the json row MUST contain a vector_layers key, whose value is an array of JSON objects.
	// Each of those JSON objects describes one layer of vector tile data, and MUST contain the following key-value pairs:
	JSON string `json:"-" mapstructure:"-"`

	//  An attribution string, which explains in English (and HTML) the sources of data and/or style for the map.
	Attribution string `json:"attribution,omitempty"`
}

// VectorLayer for style
type VectorLayer struct {
	// The layer ID, which is referred to as the name of the layer in the Mapbox Vector Tile spec.
	ID string `json:"id,omitempty" `

	// The lowest zoom level for which the tile set provides data
	MinZoom int `json:"minzoom,omitempty"`

	// The highest zoom level for which the tileset provides data
	MaxZoom int `json:"maxzoom,omitempty"`

	Fields map[string]string `json:"fields,omitempty"`
}
