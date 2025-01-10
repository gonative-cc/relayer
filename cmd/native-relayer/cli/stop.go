package cli

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the relayer",
	Run: func(cmd *cobra.Command, args []string) {
		if relayerInstance == nil {
			log.Info().Msg("Relayer is not running.")
			return
		}
		log.Info().Msg("Stopping the relayer...")
		relayerInstance.Stop()
		relayerWg.Wait()
		relayerInstance = nil
		log.Info().Msg("Relayer stopped.")

	},
}
