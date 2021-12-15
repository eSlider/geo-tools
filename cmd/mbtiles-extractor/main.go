package main

import (
	"errors"
	"log"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eslider/geo-tools/pkg/tiles"
)

// Extract tile command
var command = &cobra.Command{
	Use:                    "extract-tiles",
	Aliases:                nil,
	SuggestFor:             nil,
	Short:                  "",
	Long:                   "Extracts tiles from `mbtiles` file",
	Example:                "",
	ValidArgs:              nil,
	ValidArgsFunction:      nil,
	Args:                   cobra.NoArgs,
	ArgAliases:             nil,
	BashCompletionFunction: "",
	Deprecated:             "",
	Annotations:            nil,
	Version:                "0.0.1",
	PersistentPreRun:       nil,
	PersistentPreRunE:      nil,
	PreRun:                 nil,
	PreRunE:                nil,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(nil)
		logrus.SetFormatter(&logrus.JSONFormatter{})

		// Create exporter
		settings := tiles.ExporterSettings{
			Path:       viper.GetString("export"),
			Decompress: viper.GetBool("decompress"),
		}
		exporter, err := tiles.NewExporter(viper.GetString("import"), settings)

		if err != nil {
			logrus.WithError(err).Error("Creating tile export")
		}

		logrus.WithField("path", settings.Path).Info("Start export from")
		err = exporter.Export()
		if err != nil {
			logrus.WithError(exporter.Export()).Error("Export tiles")
		}
		logrus.Info("End export")

	},
	RunE:                       nil,
	PostRun:                    nil,
	PostRunE:                   nil,
	PersistentPostRun:          nil,
	PersistentPostRunE:         nil,
	FParseErrWhitelist:         cobra.FParseErrWhitelist{},
	TraverseChildren:           false,
	Hidden:                     false,
	SilenceErrors:              false,
	SilenceUsage:               false,
	DisableFlagParsing:         false,
	DisableAutoGenTag:          false,
	DisableFlagsInUseLine:      false,
	DisableSuggestions:         false,
	SuggestionsMinimumDistance: 0,
}

// Initializing options
func init() {
	command.Flags().StringP("import", "i", "data/countries.mbtiles", "import data path")
	command.Flags().StringP("export", "o", "tiles", "export data path")
	command.Flags().BoolP("decompress", "d", true, "decompress PBF files")
	command.Flags().BoolP("verbose", "v", true, "output extract details")
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
