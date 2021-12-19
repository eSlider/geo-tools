all: build-extractor build-geocoder

build-extractor:
	GOOS=linux \
	GOARCH=amd64 \
	CGO_ENABLED=1 \
	CGO_CFLAGS=-DSQLITE_SOUNDEX=1 \
 		go build \
 			-tags="linux osusergo netgo json1 fts5" \
 			-o dist/mbtiles-extractor \
 			cmd/mbtiles-extractor/main.go

build-geocoder:
	GOOS=linux \
	GOARCH=amd64 \
	CGO_ENABLED=1 \
	CGO_CFLAGS=-DSQLITE_SOUNDEX=1 \
 		go build \
 			-tags="linux osusergo netgo" \
 			-o dist/mbtiles-geocoder \
 			cmd/mbtiles-geocoder/main.go

clean:
	rm dist/mbtiles-*
