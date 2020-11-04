package commands

import (
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	appVersion = "0.0.1-alpha1"
)

func initLogger(stringLevel string) {
	logLevel, err := log.ParseLevel(stringLevel)
	if err != nil {
		log.Warnln("Error parsing loglevel, fallback to info")
		logLevel = log.InfoLevel
	}

	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{
		// FullTimestamp: true,
		// DisableColors: true,
	})

	log.Debugf("Log level set to %s", log.GetLevel().String())
}

// NewDefaultCommand creates the default command.
func NewDefaultCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   path.Base(os.Args[0]),
		Short: "imgsync",
		Long:  "Sync container images to a target registry based on multiple selectors and filters",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			stringLevel := viper.GetString("loglevel")
			initLogger(stringLevel)
		},
	}

	cmd.PersistentFlags().StringP("confpath", "c", "", "Dir or complete path of the config file (defaults to .imgsync.yaml)")
	viper.BindPFlag("confpath", cmd.PersistentFlags().Lookup("confpath"))

	cmd.PersistentFlags().StringP("loglevel", "l", log.InfoLevel.String(), "Log verbosity (defaults to info)")
	viper.BindPFlag("loglevel", cmd.PersistentFlags().Lookup("loglevel"))

	viper.SetEnvPrefix("IMGSYNC")
	viper.AutomaticEnv()

	cmd.AddCommand(newSyncCommand())

	return &cmd
}
