package main

import (
	"errors"
	"log"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eslider/geo-tools/pkg/mbtiles"
)

// Extract tile command
var command = &cobra.Command{
	Use:     "extract-tiles",
	Long:    "Extracts tiles from `mbtiles` file",
	Args:    cobra.NoArgs,
	Version: "0.0.1",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(nil)
		logrus.SetFormatter(&logrus.JSONFormatter{})
		if !viper.GetBool("verbose") {
			logrus.SetLevel(logrus.WarnLevel | logrus.ErrorLevel | logrus.DebugLevel | logrus.FatalLevel | logrus.PanicLevel)
		}

		// Create exporter
		settings := mbtiles.ExporterSettings{
			Path:       viper.GetString("export"),
			Decompress: viper.GetBool("decompress"),
			BaseUrl:    viper.GetString("url"),
		}

		exporter, err := mbtiles.NewExporter(viper.GetString("import"), settings)
		if err != nil {
			logrus.WithError(err).Error("Creating tile export")
		}

		logrus.WithField("path", settings.Path).Info("Start export from")
		if err = exporter.Export(); err != nil {
			logrus.WithError(exporter.Export()).Error("Export tiles")
		}
		logrus.Info("End export")
	},
}

// Initializing options
func init() {
	command.Flags().StringP("import", "i", "data/countries.mbtiles", "Import data path")
	command.Flags().StringP("export", "o", "tiles", "Export data path")
	command.Flags().StringP("url", "u", "http://localhost/tiles/", "base URL to serve tiles")
	command.Flags().BoolP("decompress", "d", true, "Decompress PBF files")
	command.Flags().BoolP("verbose", "v", false, "Output details")
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
