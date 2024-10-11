package native

import (
	"context"
	"crypto/ed25519"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	tmtypes "github.com/cometbft/cometbft/types"
)

func callMoveFunction(ctx context.Context, c *sui.Client, lcpackage string, module string, 
	function string, gasbudget string, signerAddress string, gasAddr string,
	lb *tmtypes.LightBlock) (models.TxnMetaData, error) {

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

func executeTransaction(ctx context.Context, cli *sui.Client, txnMetaData models.TxnMetaData,
	priKey ed25519.PrivateKey) (models.SuiTransactionBlockResponse, error) {

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
