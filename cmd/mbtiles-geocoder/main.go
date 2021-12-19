package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/paulmach/orb/geojson"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eslider/geo-tools/pkg/mbtiles"
)

var command = &cobra.Command{
	Use:     "mbtiles-geocoder",
	Long:    "Geocode by using mbtiles file",
	Args:    cobra.NoArgs,
	Version: "0.0.1",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(nil)
		logrus.SetFormatter(&logrus.JSONFormatter{})
		if !viper.GetBool("verbose") {
			logrus.SetLevel(logrus.WarnLevel | logrus.ErrorLevel | logrus.DebugLevel | logrus.FatalLevel | logrus.PanicLevel)
		}
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
		geoJSON, err := json.Marshal(fc)
		if err != nil {
			logrus.WithError(err).Fatal("Unable to generate GeoJSON")
		}
		fmt.Print(string(geoJSON))
	},
}

// Initializing options
func init() {
	command.Flags().BoolP("verbose", "v", false, "output details")
	command.Flags().StringP("mbtiles", "d", "data/canary-islands-latest.mbtiles", "MBtiles data path")
	command.Flags().StringP("search", "s", "", "search query")
	command.Flags().Int("max", 5, "maximal results number")
}

// main command
func main() {
	// Bind all flags
	if err := viper.BindPFlags(command.Flags()); err != nil {
		logrus.WithError(err).Fatal("Unable to bind command line flags")
	}

	// Handle environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	// Read settings from config file
	viper.AddConfigPath(".")
	viper.SetConfigName("config")

	// Get YAML
	if err := viper.ReadInConfig(); err != nil {
		// Don't fail if config not found
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			logrus.WithError(err).Warn("Unable to read config file")
		}
	}

	// Pass control
	if err := command.Execute(); err != nil {
		logrus.WithError(err).Fatal("Failed to execute command")
	}
}
