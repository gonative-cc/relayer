package main

import (
	"fmt"

	"github.com/gonative-cc/relayer/bitcoinspv"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/btcwrapper"
	"github.com/gonative-cc/relayer/lcclient"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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
			nativeClient, nativeCloser := initNativeClient(cfg)

			logTipBlock(btcClient, rootLogger)

			spvRelayer := initSPVRelayer(cfg, rootLogger, btcClient, nativeClient)
			spvRelayer.Start()

			setupShutdown(rootLogger, spvRelayer, btcClient, nativeCloser)

			<-interruptDone
			rootLogger.Info("Shutdown complete")
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", config.DefaultCfgFile(), "config file")
	return cmd
}

func initConfig(cfgFile string) (*config.Config, *zap.Logger) {
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

func initBTCClient(cfg *config.Config, rootLogger *zap.Logger) *btcwrapper.Client {
	btcClient, err := btcwrapper.NewClientWithBlockSubscriber(
		&cfg.BTC,
		cfg.Relayer.SleepDuration,
		cfg.Relayer.MaxSleepDuration,
		rootLogger,
	)
	if err != nil {
		panic(fmt.Errorf("failed to open BTC client: %w", err))
	}
	return btcClient
}

func logTipBlock(btcClient *btcwrapper.Client, rootLogger *zap.Logger) {
	latestBTCBlock, err := btcClient.GetTipBlock()
	if err != nil {
		panic(fmt.Errorf("failed to get chain tip block: %w", err))
	}

	rootLogger.Info("Got tip block",
		zap.String("hash", latestBTCBlock.Hash),
		zap.Int64("height", latestBTCBlock.Height),
		zap.Int64("time", latestBTCBlock.Time))
}

func initNativeClient(cfg *config.Config) (*lcclient.Client, func()) {
	nativeClient, nativeCloser, err := lcclient.New(cfg.Native.RPCEndpoint)
	if err != nil {
		panic(fmt.Errorf("failed to open Native client: %w", err))
	}
	return nativeClient, nativeCloser
}

func initSPVRelayer(
	cfg *config.Config,
	rootLogger *zap.Logger,
	btcClient *btcwrapper.Client,
	nativeClient *lcclient.Client,
) *bitcoinspv.Relayer {
	spvRelayer, err := bitcoinspv.New(
		&cfg.Relayer,
		rootLogger,
		btcClient,
		nativeClient,
		cfg.Relayer.SleepDuration,
		cfg.Relayer.MaxSleepDuration,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create bitcoin-spv relayer: %w", err))
	}
	return spvRelayer
}

func setupShutdown(
	rootLogger *zap.Logger,
	spvRelayer *bitcoinspv.Relayer,
	btcClient *btcwrapper.Client,
	nativeCloser func(),
) {
	registerHandler(func() {
		rootLogger.Info("Stopping relayer...")
		spvRelayer.Stop()
		rootLogger.Info("Relayer shutdown")
	})
	registerHandler(func() {
		rootLogger.Info("Stopping BTC client...")
		btcClient.Stop()
		btcClient.WaitForShutdown()
		rootLogger.Info("BTC client shutdown")
	})
	registerHandler(func() {
		rootLogger.Info("Stopping Native client...")
		nativeCloser()
		rootLogger.Info("Native client shutdown")
	})
}
