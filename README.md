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

### Configuration options

#### Import

* `--import.path`: Export database file path and is `data/tiles-world-vector.mbtiles` by default.

#### Export

* `--path`, `-i`: Export path which is `tiles` by default.
* `--decompress`, `-d`: Determinate tile compression format and export raw `PBF` tiles.
* `--url`, `-u`: base URL to serve tiles (default `http://localhost/tiles/`)
