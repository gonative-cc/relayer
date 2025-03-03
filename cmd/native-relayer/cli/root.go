package cli

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	configFile string
	logLevel   string

	rootCmd = &cobra.Command{
		Use:   "relayer-cli",
		Short: "CLI tool for managing the relayer",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("CLI execution failed")
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to the config file")
	rootCmd.PersistentFlags().StringVar(
		&logLevel, "log-level",
		"",
		"Set the log level (trace, debug, info, warn, error, fatal, panic)",
	)
	rootCmd.AddCommand(startCmd)
}
