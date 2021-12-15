# Geo Tools

A set of tools to transform or deal with geo data.

## Installation

```shell
git clone git@github.com/eSlider/geo-tools
go mod download
make
```

## mbtiles to PBF tile extractor

Blasting fast mbtiles to PBF tile extractor written in golang.

It can export `pbf`, `jpeg`, `webp` or `png` files into `/z/y/x/[number].[pbf|webp|png|jpg]` file structure from an `mbtiles` map database file.

### Run example

```shell
dist/mbtiles-extractor -i data/tiles-world-vector.mbtiles
```

### Configuration options

#### Import

* `--import.path`: Export database file path and is `data/tiles-world-vector.mbtiles` by default.

#### Export

* `--path` or `-i`: Export path which is `tiles` by default.
* `--decompress` or `-d`: Determinate tile compression format and export raw `PBF` tiles.

