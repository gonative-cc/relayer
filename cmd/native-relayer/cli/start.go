package cli

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/env"
	"github.com/gonative-cc/relayer/ika"
	"github.com/gonative-cc/relayer/ika2btc"
	"github.com/gonative-cc/relayer/native"
	"github.com/gonative-cc/relayer/native2ika"
	"github.com/gonative-cc/relayer/nbtc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the relayer",
	Run: func(cmd *cobra.Command, args []string) {
		lvl, err := cmd.Root().PersistentFlags().GetString("log-level")
		if err != nil {
			log.Error().Err(err).Msg("Error getting log level")
			os.Exit(1)
		}
		logLvl, err := zerolog.ParseLevel(lvl)
		if err != nil {
			log.Error().Err(err).Msg("Error parsing log level")
			os.Exit(1)
		}
		env.InitLogger(logLvl)
		configFile, err := cmd.Root().PersistentFlags().GetString("config")
		if err != nil {
			log.Error().Err(err).Msg("Error getting config file path")
			os.Exit(1)
		}

		config, err := loadConfig(configFile)
		if err != nil {
			log.Error().Err(err).Msg("Error loading config")
			os.Exit(1)
		}
		db, err := dal.NewDB(config.DB.File)
		if err != nil {
			log.Error().Err(err).Msg("Error creating database")
			os.Exit(1)
		}
		if err := db.InitDB(); err != nil {
			log.Error().Err(err).Msg("Error initializing database")
			os.Exit(1)
		}
		suiClient := sui.NewSuiClient(config.Ika.RPC).(*sui.Client)
		if suiClient == nil {
			log.Error().Err(err).Msg("Error creating Sui client")
			os.Exit(1)
		}
		ikaSigner, err := signer.NewSignertWithMnemonic(config.Ika.SignerMnemonic)
		if err != nil {
			log.Error().Err(err).Msg("Error creating signer with mnemonic")
			os.Exit(1)
		}
		ikaClient, err := ika.NewClient(
			suiClient,
			ikaSigner,
			ika.SuiCtrCall{
				Package:  config.Ika.NativeLcPackage,
				Module:   config.Ika.NativeLcModule,
				Function: config.Ika.NativeLcFunction,
			},
			ika.SuiCtrCall{
				Package:  config.Ika.NativeLcPackage,
				Module:   config.Ika.NativeLcModule,
				Function: config.Ika.NativeLcFunction,
			},
			config.Ika.GasAcc,
			config.Ika.GasBudget,
		)
		if err != nil {
			log.Error().Err(err).Msg("Error creating IKA client")
			os.Exit(1)
		}
		btcProcessor, err := ika2btc.NewProcessor(
			rpcclient.ConnConfig{
				Host:         config.Btc.RPCHost,
				User:         config.Btc.RPCUser,
				Pass:         config.Btc.RPCPass,
				HTTPPostMode: config.Btc.HTTPPostMode,
				DisableTLS:   config.Btc.DisableTLS,
			},
			config.Btc.ConfirmationThreshold,
			db,
		)
		if err != nil {
			log.Error().Err(err).Msg("Error creating Bitcoin processor")
			os.Exit(1)
		}
		nativeProcessor := native2ika.NewProcessor(ikaClient, db)
		// TODO: replace with the the real endpoint  once available
		fetcher, err := native.NewMockAPISignRequestFetcher()
		if err != nil {
			log.Error().Err(err).Msg("Error creating SignReq fetcher ")
			os.Exit(1)
		}
		relayer, err := nbtc.NewRelayer(
			nbtc.RelayerConfig{
				ProcessTxsInterval:   config.Relayer.ProcessTxsInterval,
				ConfirmTxsInterval:   config.Relayer.ConfirmTxsInterval,
				SignReqFetchInterval: config.Relayer.SignReqFetchInterval,
				SignReqFetchFrom:     config.Relayer.SignReqFetchFrom,
				SignReqFetchLimit:    config.Relayer.SignReqFetchLimit,
			},
			db,
			nativeProcessor,
			btcProcessor,
			fetcher,
		)
		if err != nil {
			log.Error().Err(err).Msg("Error creating relayer")
			os.Exit(1)
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := relayer.Start(context.Background()); err != nil {
				log.Error().Err(err).Msg("Relayer encountered an error")
			}
		}()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		log.Info().Msg("Stopping the relayer...")
		relayer.Stop()
		wg.Wait()
		log.Info().Msg("Relayer stopped.")
	},
}
