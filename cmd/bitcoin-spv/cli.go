package main

import (
	"fmt"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/gonative-cc/relayer/bitcoinspv"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/btcclient"
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
			var (
				err          error
				cfg          config.Config
				btcClient    *btcclient.Client
				nativeClient *lcclient.Client
				spvRelayer   *bitcoinspv.Relayer
				nativeCloser jsonrpc.ClientCloser
				// server           *rpcserver.Server
			)

			// get the config from the given file or the default file
			cfg, err = config.New(cfgFile)
			if err != nil {
				panic(fmt.Errorf("failed to load config: %w", err))
			}
			rootLogger, err := cfg.CreateLogger()
			if err != nil {
				panic(fmt.Errorf("failed to create logger: %w", err))
			}

			// create BTC client and connect to BTC server
			// Note that bitcoin-spv relayer needs to subscribe to new BTC blocks
			btcClient, err = btcclient.NewWithBlockSubscriber(
				&cfg.BTC,
				cfg.Common.RetrySleepTime,
				cfg.Common.MaxRetrySleepTime,
				rootLogger,
			)
			if err != nil {
				panic(fmt.Errorf("failed to open BTC client: %w", err))
			}

			// get tip block info
			tipBlock, err := btcClient.GetTipBlockVerbose()
			if err != nil {
				panic(fmt.Errorf("failed to get chain tip block: %w", err))
			}

			rootLogger.Info("got tip block",
				zap.String("hash", tipBlock.Hash),
				zap.Int64("height", tipBlock.Height),
				zap.Int64("time", tipBlock.Time))

			// create Native client. Note that requests from Native client are ad hoc
			nativeClient, nativeCloser, err = lcclient.New("http://localhost:9797")
			if err != nil {
				panic(fmt.Errorf("failed to open Native client: %w", err))
			}

			// register relayer metrics
			relayerMetrics := bitcoinspv.NewRelayerMetrics()

			// create relayer
			spvRelayer, err = bitcoinspv.New(
				&cfg.Relayer,
				rootLogger,
				btcClient,
				nativeClient,
				cfg.Common.RetrySleepTime,
				cfg.Common.MaxRetrySleepTime,
				relayerMetrics,
			)
			if err != nil {
				panic(fmt.Errorf("failed to create bitcoin-spv relayer: %w", err))
			}

			// start normal-case execution
			spvRelayer.Start()

			// // start Prometheus metrics server
			// addr := fmt.Sprintf("%s:%d", cfg.Metrics.Host, cfg.Metrics.ServerPort)
			// spvRelayer.MetricsStart(addr, relayerMetrics.Registry)

			// TODO: uncomment before pushing
			// SIGINT handling stuff
			addInterruptHandler(func() {
				rootLogger.Info("Stopping relayer...")
				spvRelayer.Stop()
				rootLogger.Info("Relayer shutdown")
			})
			addInterruptHandler(func() {
				rootLogger.Info("Stopping BTC client...")
				btcClient.Stop()
				btcClient.WaitForShutdown()
				rootLogger.Info("BTC client shutdown")
			})
			addInterruptHandler(func() {
				rootLogger.Info("Stopping Native client...")
				nativeCloser()
				rootLogger.Info("Native client shutdown")
			})

			<-interruptHandlersDone
			rootLogger.Info("Shutdown complete")
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", config.DefaultConfigFile(), "config file")
	return cmd
}
