package main

import (
	"fmt"

	"github.com/gonative-cc/relayer/bitcoinspv"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/clients/btcwrapper"
	"github.com/gonative-cc/relayer/bitcoinspv/clients/sui"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/pattonkan/sui-go/suiclient"
	"github.com/pattonkan/sui-go/suisigner"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "bitcoin-spv",
		Short: "An relayer for Bitcoin blocks -> Native",
	}
)

func init() {
	rootCmd.AddCommand(CmdStart())
}

// CmdExecute executes the root command.
func CmdExecute() error {
	return rootCmd.Execute()
}

// CmdStart returns the CLI commands for the bitcoin-spv
func CmdStart() *cobra.Command {
	var cfgFile = ""
	var storeInWalrus = false

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the bitcoin-spv relayer",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, rootLogger := initConfig(cfgFile)
			if storeInWalrus {
				cfg.Relayer.StoreBlocksInWalrus = true
			}

			btcClient := initBTCClient(cfg, rootLogger)
			nativeClient := initNativeClient(cfg, rootLogger)
			walrusHandler, err := initWalrusHandler(&cfg.Relayer, rootLogger) // will return nil if flag not set
			if err != nil {
				return err
			}

			logTipBlock(btcClient, rootLogger)

			spvRelayer := initSPVRelayer(cfg, rootLogger, btcClient, nativeClient, walrusHandler)
			spvRelayer.Start()

			setupShutdown(rootLogger, spvRelayer, btcClient, nativeClient)

			<-interruptDone
			rootLogger.Info().Msg("Shutdown complete")
			return nil
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", config.DefaultCfgFile(), "config file")
	cmd.Flags().BoolVar(&storeInWalrus, "walrus", false, "enable storing full blocks in Walrus")
	return cmd
}

func initConfig(cfgFile string) (*config.Config, zerolog.Logger) {
	cfg, err := config.New(cfgFile)
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}
	rootLogger, err := cfg.CreateLogger()
	if err != nil {
		panic(fmt.Errorf("failed to create logger: %w", err))
	}
	return &cfg, rootLogger
}

func initBTCClient(cfg *config.Config, rootLogger zerolog.Logger) *btcwrapper.Client {
	btcClient, err := btcwrapper.NewClientWithBlockSubscriber(
		&cfg.BTC,
		cfg.Relayer.RetrySleepDuration,
		cfg.Relayer.MaxRetrySleepDuration,
		rootLogger,
	)
	if err != nil {
		panic(fmt.Errorf("failed to open BTC client: %w", err))
	}
	return btcClient
}

func logTipBlock(btcClient *btcwrapper.Client, rootLogger zerolog.Logger) {
	latestBTCBlock, err := btcClient.GetTipBlock()
	if err != nil {
		panic(fmt.Errorf("failed to get chain tip block: %w", err))
	}

	rootLogger.Info().
		Str("hash", latestBTCBlock.Hash).
		Int64("height", latestBTCBlock.Height).
		Int64("time", latestBTCBlock.Time).
		Msg("Got tip block")
}

func initNativeClient(cfg *config.Config, rootLogger zerolog.Logger) clients.BitcoinSPV {
	c := suiclient.NewClient(cfg.Sui.Endpoint)

	signer, err := suisigner.NewSignerWithMnemonic(cfg.Sui.Mnemonic, suisigner.KeySchemeFlagDefault)
	if err != nil {
		panic(fmt.Errorf("failed to create new signer: %w", err))
	}

	client, err := sui.New(c, signer, cfg.Sui.LCObjectID, cfg.Sui.LCPackageID, rootLogger)

	if err != nil {
		panic(fmt.Errorf("failed to create new bitcoinSPVClient: %w", err))
	}

	return client
}

func initWalrusHandler(cfg *config.RelayerConfig, rootLogger zerolog.Logger) (*bitcoinspv.WalrusHandler, error) {
	wh, err := bitcoinspv.NewWalrusHandler(cfg, rootLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize WalrusHandler: %w", err)
	}
	return wh, nil
}

func initSPVRelayer(
	cfg *config.Config,
	rootLogger zerolog.Logger,
	btcClient *btcwrapper.Client,
	nativeClient clients.BitcoinSPV,
	walrusHandler *bitcoinspv.WalrusHandler,
) *bitcoinspv.Relayer {
	spvRelayer, err := bitcoinspv.New(
		&cfg.Relayer,
		rootLogger,
		btcClient,
		nativeClient,
		walrusHandler,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create bitcoin-spv relayer: %w", err))
	}
	return spvRelayer
}

// Shutdown relayer
func setupShutdown(
	rootLogger zerolog.Logger,
	spvRelayer *bitcoinspv.Relayer,
	btcClient *btcwrapper.Client,
	nativeClient clients.BitcoinSPV,
) {
	registerHandler(func() {
		rootLogger.Info().Msg("Stopping relayer...")
		spvRelayer.Stop()
		rootLogger.Info().Msg("Relayer shutdown")
	})
	registerHandler(func() {
		rootLogger.Info().Msg("Stopping BTC client...")
		btcClient.Stop()
		btcClient.WaitForShutdown()
		rootLogger.Info().Msg("BTC client shutdown")
	})
	registerHandler(func() {
		rootLogger.Info().Msg("Stopping Native client...")
		nativeClient.Stop()
		rootLogger.Info().Msg("Native client shutdown")
	})
}
