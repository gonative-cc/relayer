package main

import (
	"fmt"

	"github.com/filecoin-project/go-jsonrpc"
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
			var (
				err          error
				cfg          config.Config
				btcClient    *btcwrapper.Client
				nativeClient *lcclient.Client
				spvRelayer   *bitcoinspv.Relayer
				nativeCloser jsonrpc.ClientCloser
			)

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
			btcClient, err = btcwrapper.NewClientWithBlockSubscriber(
				&cfg.BTC,
				cfg.Relayer.SleepDuration,
				cfg.Relayer.MaxSleepDuration,
				rootLogger,
			)
			if err != nil {
				panic(fmt.Errorf("failed to open BTC client: %w", err))
			}

			tipBlock, err := btcClient.GetTipBlock()
			if err != nil {
				panic(fmt.Errorf("failed to get chain tip block: %w", err))
			}

			rootLogger.Info("got tip block",
				zap.String("hash", tipBlock.Hash),
				zap.Int64("height", tipBlock.Height),
				zap.Int64("time", tipBlock.Time))

			nativeClient, nativeCloser, err = lcclient.New(cfg.Native.RPCEndpoint)
			if err != nil {
				panic(fmt.Errorf("failed to open Native client: %w", err))
			}

			spvRelayer, err = bitcoinspv.New(
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

			// start normal-case execution
			spvRelayer.Start()

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
