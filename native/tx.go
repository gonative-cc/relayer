package native

import (
	"context"
	"crypto/ed25519"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	tmtypes "github.com/cometbft/cometbft/types"
)

// callMoveFunction prepares a Move call transaction to be executed on the Ika/Sui network.
// It takes the necessary parameters for the Move call, including the package, module,
// function, arguments, gas information, and signer address.
// It returns the transaction metadata and an error if any occurred during preparation.
func callMoveFunction(
	ctx context.Context,
	c *sui.Client,
	lcpackage, module, function, gasbudget,
	signerAddress, gasAddr string,
	lb *tmtypes.LightBlock,
) (models.TxnMetaData, error) {
	return c.MoveCall(ctx, models.MoveCallRequest{
		Signer:          signerAddress,
		PackageObjectId: lcpackage,
		Module:          module,
		Function:        function,
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			lb,
		},
		Gas:       gasAddr,
		GasBudget: gasbudget,
	})
}

// executeTransaction signs and executes a prepared Move call transaction on the Ika/Sui network.
// It takes the transaction metadata, the signer's private key, and execution options.
// It returns the transaction response and an error if any occurred during execution.
func executeTransaction(
	ctx context.Context,
	cli *sui.Client,
	txnMetaData models.TxnMetaData,
	priKey ed25519.PrivateKey,
) (models.SuiTransactionBlockResponse, error) {
	return cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: txnMetaData,
		PriKey:      priKey,
		Options: models.SuiTransactionBlockOptions{
			ShowInput:    true,
			ShowRawInput: true,
			ShowEffects:  true,
		},
		RequestType: "WaitForLocalExecution",
	})
}
