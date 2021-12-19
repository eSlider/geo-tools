# Geo Tools

A set of tools to transform or deal with geo data.

## Installation

```shell
git clone git@github.com/eSlider/geo-tools
go mod download
make
```

## MBTiles to PBF tile extractor

Blasting fast MBTiles to PBF tile extractor written in golang.

It can export `pbf`, `jpeg`, `webp` or `png` files into `/z/y/x/[number].[pbf|webp|png|jpg]` file structure from an `mbtiles` map database file.

### Run example

```shell
dist/mbtiles-extractor -i data/tiles-world-vector.mbtiles -u http://localhost/tiles
```

### Flags

#### Import

* `--import.path`: Export database file path and is `data/tiles-world-vector.mbtiles` by default.

#### Export

* `--path`, `-i`: Export path which is `tiles` by default.
* `--decompress`, `-d`: Determinate tile compression format and export raw `PBF` tiles.
* `--url`, `-u`: base URL to serve tiles (default `http://localhost/tiles/`)

## Geocode by using mbtiles file

### Run example

Get all features from mbtiles file as GeoJSON:

```shell
dist/mbtiles-geocoder -d data/canary-islands-latest.mbtiles -s "" --max 1000
```

### Flags

* `-d`, `--mbtiles` `string`: MBtiles data path (default `data/canary-islands-latest.mbtiles`)
* `-s`, `--search` `string`: search query
* `--max` `int`: maximal results number (default `5`)

## Issues

There some operating system limits can be turned off before run concurrent exporting:

```shell
ulimit -s unlimited
```

