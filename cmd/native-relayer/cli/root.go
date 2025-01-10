package cli

import (
	"os"
	"sync"

	"github.com/gonative-cc/relayer/nbtc"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	verbose bool // Flag for verbose mode

	rootCmd = &cobra.Command{
		Use:   "relayer-cli",
		Short: "CLI tool for managing the relayer",
	}

	relayerInstance *nbtc.Relayer
	relayerWg       sync.WaitGroup
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("CLI execution failed")
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
}
