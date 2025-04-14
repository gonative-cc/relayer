package main

import (
	"fmt"

	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/gonative-cc/relayer/bitcoinspv"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/clients/btcwrapper"
	suiClient "github.com/gonative-cc/relayer/bitcoinspv/clients/sui"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
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

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the bitcoin-spv relayer",
		Run: func(_ *cobra.Command, _ []string) {
			cfg, rootLogger := initConfig(cfgFile)
			btcClient := initBTCClient(cfg, rootLogger)
			spvClient := initSPVClient(cfg, rootLogger)

			// TODO: do we need to log it here? Why we panic inside of it
			logTipBlock(btcClient, rootLogger)

			spvRelayer := initSPVRelayer(cfg, rootLogger, btcClient, spvClient)
			spvRelayer.Start()

			setupShutdown(rootLogger, spvRelayer, btcClient, spvClient)

			<-interruptDone
			rootLogger.Info().Msg("Shutdown complete")
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", config.DefaultCfgFile(), "config file")
	return cmd
}

// TODO: rename to mustInitConfig or return the error and handle it in the CmdStart
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

// TODO: same as above
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

// TODO: consider rename or return error
func initSPVClient(cfg *config.Config, rootLogger zerolog.Logger) clients.BitcoinSPV {
	c := sui.NewSuiClient(cfg.Sui.Endpoint).(*sui.Client)

	signer, err := signer.NewSignertWithMnemonic(cfg.Sui.Mnemonic)
	if err != nil {
		panic(fmt.Errorf("failed to create new signer: %w", err))
	}

	client, err := suiClient.NewSPVClient(c, signer, cfg.Sui.LCObjectID, cfg.Sui.LCPackageID, rootLogger)

	if err != nil {
		panic(fmt.Errorf("failed to create new bitcoinSPVClient: %w", err))
	}

	return client
}

// TODO: consider returning error
func initSPVRelayer(
	cfg *config.Config,
	rootLogger zerolog.Logger,
	btcClient *btcwrapper.Client, //TODO: consider converting this to clients.BTCClient interface
	nativeClient clients.BitcoinSPV,
) *bitcoinspv.Relayer { //TODO: consider interface
	spvRelayer, err := bitcoinspv.New(
		&cfg.Relayer,
		rootLogger,
		btcClient,
		nativeClient,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create bitcoin-spv relayer: %w", err))
	}
	return spvRelayer
}

// TODO: do we need to shutdown spvClient (sui)
func setupShutdown(
	rootLogger zerolog.Logger,
	spvRelayer *bitcoinspv.Relayer, //TODO: consider converting this to interface
	btcClient *btcwrapper.Client, //TODO: consider converting this to clients.BTCClient interface
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
