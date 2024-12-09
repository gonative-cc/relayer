package main

import (
	"fmt"
	"os"

	bbnclient "github.com/babylonchain/babylon/client/client"
	"github.com/gonative-cc/relayer/btcclient"
	"github.com/gonative-cc/relayer/reporter"
	"github.com/gonative-cc/relayer/reporter/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// ENV variables
const (
	EnvChainRPC            = "NATIVE_RPC"
	EnvChainGRPC           = "NATIVE_GRPC"
	FlagMinimumBlockHeight = "block"
	defaultPort            = "7272"
)

var (
	rootCmd = &cobra.Command{
		Use:   "reporter",
		Short: "An relayer for Bitcoin blocks -> Native",
	}
)

func init() {
	rootCmd.AddCommand(CmdStart())
}

// CmdExecute executes the root command.
func CmdExecute() error {
	printEnv()
	return rootCmd.Execute()
}

// CmdStart returns the CLI commands for the reporter
func CmdStart() *cobra.Command {
	var babylonKeyDir string
	var cfgFile = ""

	cmd := &cobra.Command{
		Use:   "reporter",
		Short: "Vigilant reporter",
		Run: func(_ *cobra.Command, _ []string) {
			var (
				err              error
				cfg              config.Config
				btcClient        *btcclient.Client
				babylonClient    *bbnclient.Client
				vigilantReporter *reporter.Reporter
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

			// apply the flags from CLI
			if len(babylonKeyDir) != 0 {
				cfg.Babylon.KeyDirectory = babylonKeyDir
			}

			// create BTC client and connect to BTC server
			// Note that vigilant reporter needs to subscribe to new BTC blocks
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
			} else {
				rootLogger.Info("got tip block",
					zap.String("hash", tipBlock.Hash),
					zap.Int64("height", tipBlock.Height),
					zap.Int64("time", tipBlock.Time))
			}

			// create Babylon client. Note that requests from Babylon client are ad hoc
			babylonClient, err = bbnclient.New(&cfg.Babylon, nil)
			if err != nil {
				panic(fmt.Errorf("failed to open Babylon client: %w", err))
			}

			// register reporter metrics
			reporterMetrics := reporter.NewReporterMetrics()

			// create reporter
			vigilantReporter, err = reporter.New(
				&cfg.Reporter,
				rootLogger,
				btcClient,
				babylonClient,
				cfg.Common.RetrySleepTime,
				cfg.Common.MaxRetrySleepTime,
				reporterMetrics,
			)
			if err != nil {
				panic(fmt.Errorf("failed to create vigilante reporter: %w", err))
			}

			// // create RPC server
			// server, err = rpcserver.New(&cfg.GRPC, rootLogger, nil, vigilantReporter, nil, nil)
			// if err != nil {
			// 	panic(fmt.Errorf("failed to create reporter's RPC server: %w", err))
			// }

			// start normal-case execution
			vigilantReporter.Start()

			// // start RPC server
			// server.Start()

			// // start Prometheus metrics server
			// addr := fmt.Sprintf("%s:%d", cfg.Metrics.Host, cfg.Metrics.ServerPort)
			// reporter.MetricsStart(addr, reporterMetrics.Registry)

			// TODO: uncomment before pushing
			// // SIGINT handling stuff
			// addInterruptHandler(func() {
			// 	// TODO: Does this need to wait for the grpc server to finish up any requests?
			// 	rootLogger.Info("Stopping RPC server...")
			// 	server.Stop()
			// 	rootLogger.Info("RPC server shutdown")
			// })
			// addInterruptHandler(func() {
			// 	rootLogger.Info("Stopping reporter...")
			// 	vigilantReporter.Stop()
			// 	rootLogger.Info("Reporter shutdown")
			// })
			// addInterruptHandler(func() {
			// 	rootLogger.Info("Stopping BTC client...")
			// 	btcClient.Stop()
			// 	btcClient.WaitForShutdown()
			// 	rootLogger.Info("BTC client shutdown")
			// })

			// <-interruptHandlersDone
			// rootLogger.Info("Shutdown complete")

		},
	}
	cmd.Flags().StringVar(&babylonKeyDir, "babylon-key-dir", "", "Directory of the Babylon key")
	cmd.Flags().StringVar(&cfgFile, "config", config.DefaultConfigFile(), "config file")
	return cmd
}

// just prints the env file.
func printEnv() {
	fmt.Printf(
		"__ENVS used__\n%s = %s\n%s = %s\n-----------------\n",
		EnvChainRPC, os.Getenv(EnvChainRPC),
		EnvChainGRPC, os.Getenv(EnvChainGRPC),
	)
}
