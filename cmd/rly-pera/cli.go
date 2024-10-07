package main

import (
	"fmt"
	"os"

	"github.com/gonative-cc/relayer/native"
	"github.com/gonative-cc/relayer/native/blockchain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ENV variables
const (
	EnvChainRPC            = "NATIVE_RPC"
	EnvChainGRPC           = "NATIVE_GRPC"
	FlagMinimumBlockHeight = "block"
	defaultPort            = "8080"
	SuiChain               = "SUI_CHAIN"
)

var (
	rootCmd = &cobra.Command{
		Use:   "rly-pera",
		Short: "An relayer for Native <-> Pera MPC",
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

// CmdStart starts the daemon to listen for new blocks
func CmdStart() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the relayer, querying and listening for the new blocks.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			minimumBlockHeight, err := cmd.Flags().GetInt(FlagMinimumBlockHeight)
			if err != nil {
				return err
			}
			// TODO: load the (should latest indexed block height from a file)
			log.Info().Int("block", minimumBlockHeight).Msg("Start relaying msgs")

			b, err := blockchain.New(os.Getenv(EnvChainRPC), os.Getenv(EnvChainGRPC))
			if err != nil {
				return err
			}
			cli, err := native.CreateSuiClient(os.Getenv(SuiChain))
			if err != nil {
				return err
			}

			logger := log.With().Str("module", "native").Logger()
			ctx := cmd.Context()

			idx, err := native.NewIndexer(ctx, b, logger, minimumBlockHeight, cli)
			if err != nil {
				return err
			}
			return idx.Start(ctx)
		},
	}

	cmd.Flags().Int(FlagMinimumBlockHeight, 1, fmt.Sprintf(
		"%s=100 to start relaying from block 100", FlagMinimumBlockHeight))
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
