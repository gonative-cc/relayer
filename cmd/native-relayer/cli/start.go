package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/ika"
	"github.com/gonative-cc/relayer/ika2btc"
	"github.com/gonative-cc/relayer/native2ika"
	"github.com/gonative-cc/relayer/nbtc"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the relayer",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			log.Error().Err(err).Msg("Error loading config")
			os.Exit(1)
		}

		// TODO:Set up logging based on verbose flag

		db, err := dal.NewDB(config.DBFile)
		if err != nil {
			log.Error().Err(err).Msg("Error creating database")
			os.Exit(1)
		}
		if err := db.InitDB(); err != nil {
			log.Error().Err(err).Msg("Error initializing database")
			os.Exit(1)
		}

		suiClient := sui.NewSuiClient(config.IkaRPC).(*sui.Client)
		if suiClient == nil {
			log.Error().Err(err).Msg("Error creating Sui client")
			os.Exit(1)
		}
		signer, err := signer.NewSignertWithMnemonic(config.IkaSignerMnemonic)
		if err != nil {
			log.Error().Err(err).Msg("Error creating signer with mnemonic")
			os.Exit(1)
		}

		ikaClient, err := ika.NewClient(
			suiClient,
			signer,
			ika.SuiCtrCall{
				Package:  config.IkaNativeLcPackage,
				Module:   config.IkaNativeLcModule,
				Function: config.IkaNativeLcFunction,
			},
			ika.SuiCtrCall{
				Package:  config.IkaNativeLcPackage,
				Module:   config.IkaNativeLcModule,
				Function: config.IkaNativeLcFunction,
			},
			config.IkaGasAcc,
			config.IkaGasBudget,
		)
		if err != nil {
			log.Error().Err(err).Msg("Error creating IKA client")
			os.Exit(1)
		}

		btcProcessor, err := ika2btc.NewProcessor(
			rpcclient.ConnConfig{
				Host:         config.BtcRPCHost,
				User:         config.BtcRPCUser,
				Pass:         config.BtcRPCPass,
				HTTPPostMode: true,
				DisableTLS:   false,
			},
			config.BtcConfirmationThreshold,
			db,
		)
		if err != nil {
			log.Error().Err(err).Msg("Error creating Bitcoin processor")
			os.Exit(1)
		}
		btcProcessor.BtcClient = &bitcoin.MockClient{}

		nativeProcessor := native2ika.NewProcessor(ikaClient, db)

		relayer, err := nbtc.NewRelayer(
			nbtc.RelayerConfig{
				ProcessTxsInterval:    config.ProcessTxsInterval,
				ConfirmTxsInterval:    config.ConfirmTxsInterval,
				ConfirmationThreshold: config.BtcConfirmationThreshold,
			},
			db, nativeProcessor,
			btcProcessor)
		if err != nil {
			log.Error().Err(err).Msg("Error creating relayer")
			os.Exit(1)
		}

		relayerWg.Add(1)
		go func() {
			defer relayerWg.Done()
			if err := relayer.Start(context.Background()); err != nil {
				log.Error().Err(err).Msg("Relayer encountered an error")
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
