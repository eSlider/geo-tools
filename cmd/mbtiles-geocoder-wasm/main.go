//go:build js && wasm
// +build js,wasm

package main

//go:generate cp $GOROOT/misc/wasm/wasm_exec.js .

import (
	"encoding/json"

	"github.com/gopherjs/gopherjs/js"
	"github.com/paulmach/orb/geojson"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/eslider/geo-tools/pkg/mbtiles"
)

// main command
func main() {
	c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// register functions
	js.Global().Set("search", js.MakeFunc(func(i []js.Value) {
		js.Global().Set("output", js.ValueOf(i[0].Int()+i[1].Int()))
	}))
	<-c
}

func search(db string, subject string, maxResults string) ([]byte, error) {

	manager, err := mbtiles.NewManager(viper.GetString("mbtiles"))
	if err != nil {
		logrus.WithError(err).Fatal("Unable to open mbtiles database")
	}

	features, err := manager.Search(viper.GetString("search"), viper.GetInt("max"))
	if err != nil {
		logrus.WithError(err).Fatal("Unable to search database")
	}

	fc := &geojson.FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	return json.Marshal(fc)
}
