package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/ika"
	"github.com/gonative-cc/relayer/ika2btc"
	"github.com/gonative-cc/relayer/native2ika"
	"github.com/gonative-cc/relayer/nbtc"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the relayer",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			// TODO:Handle error
		}

		// TODO:Set up logging based on verbose flag

		db, err := dal.NewDB(":memory:")
		if err != nil {
			// TODO:Handle error
		}
		if err := db.InitDB(); err != nil {
			// TODO:Handle error
		}

		mockIkaClient := ika.NewMockClient()

		btcProcessor, err := ika2btc.NewProcessor(
			rpcclient.ConnConfig{
				Host:         "test_rpc",
				User:         "test_user",
				Pass:         "test_pass",
				HTTPPostMode: true,
				DisableTLS:   false,
			},
			6,
			db,
		)
		if err != nil {
			// TODO:Handle error
		}
		btcProcessor.BtcClient = &bitcoin.MockClient{}

		nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)

		relayer, err := nbtc.NewRelayer(*config, db, nativeProcessor, btcProcessor)
		if err != nil {
			// TODO:Handle error
		}

		relayerWg.Add(1)
		go func() {
			defer relayerWg.Done()
			if err := relayer.Start(context.Background()); err != nil {
				// TODO:Handle error
			}
		}()

		relayerInstance = relayer

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		fmt.Println("Stopping the relayer...")
		relayerInstance.Stop()
		relayerWg.Wait()
		fmt.Println("Relayer stopped.")
	},
}
