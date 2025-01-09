package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the relayer",
	Run: func(cmd *cobra.Command, args []string) {
		if relayerInstance == nil {
			fmt.Println("Relayer is not running.")
			return
		}
		fmt.Println("Stopping the relayer...")
		relayerInstance.Stop()
		relayerWg.Wait()
		relayerInstance = nil
		fmt.Println("Relayer stopped.")
	},
}
