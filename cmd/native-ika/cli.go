package main

import (
	"fmt"
	"os"

	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
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
	IkaChain               = "IKA_RPC"
	IkaSignerMnemonic      = "IKA_SIGNER_MNEMONIC"
	IkaNativeLcPackage     = "IKA_NATIVE_LC_PACKAGE"
	IkaNativeLcModule      = "IKA_NATIVE_LC_MODULE"
	IkaNativeLcFunction    = "IKA_NATIVE_LC_FUNCTION"
	GasAddress             = "GAS_ADDRESS"
	GasBudget              = "GAS_BUDGET"
)

var (
	rootCmd = &cobra.Command{
		Use:   "rly-ika",
		Short: "A relayer for Native <-> Ika MPC",
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
			c := sui.NewSuiClient(os.Getenv(IkaChain)).(*sui.Client)

			signer, err := signer.NewSignertWithMnemonic(os.Getenv(IkaSignerMnemonic))
			if err != nil {
				return err
			}

			logger := log.With().Str("module", "native").Logger()
			ctx := cmd.Context()

			pc, err := native.NewIkaClient(c, signer,
				os.Getenv(IkaNativeLcPackage), os.Getenv(IkaNativeLcModule),
				os.Getenv(IkaNativeLcFunction), os.Getenv(GasAddress), os.Getenv(GasBudget))
			if err != nil {
				return err
			}

			idx, err := native.NewIndexer(ctx, b, logger, minimumBlockHeight, pc)
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
		"__ENVS used__\n%s = %s\n%s = %s\n%s = %s\n%s = %s\n%s = %s\n%s = %s\n%s = %s\n%s = %s\n%s = %s\n-----------------\n",
		EnvChainRPC, os.Getenv(EnvChainRPC),
		EnvChainGRPC, os.Getenv(EnvChainGRPC),
		IkaChain, os.Getenv(IkaChain),
		IkaSignerMnemonic, os.Getenv(IkaSignerMnemonic),
		IkaNativeLcPackage, os.Getenv(IkaNativeLcPackage),
		IkaNativeLcModule, os.Getenv(IkaNativeLcModule),
		IkaNativeLcFunction, os.Getenv(IkaNativeLcFunction),
		GasAddress, os.Getenv(GasAddress),
		GasBudget, os.Getenv(GasBudget),
	)
}
