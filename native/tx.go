package native

import (
	"context"
	"crypto/ed25519"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	tmtypes "github.com/cometbft/cometbft/types"
)




func callMoveFunction(ctx context.Context, cli *sui.Client, signerAddress string, gasObj string, lb *tmtypes.LightBlock) (models.TxnMetaData, error) {
	return cli.MoveCall(ctx, models.MoveCallRequest{
		Signer:          signerAddress,
		PackageObjectId: "0x97436bc2b3bba89b96dee5288e43dd96472c0e7d20e4ea8bd36c7ff636771ee9",
		Module:          "example",
		Function:        "magic",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			lb,
		},
		Gas:       gasObj,
		GasBudget: "100000000",
	})
}

func executeTransaction(ctx context.Context, cli *sui.Client, txnMetaData models.TxnMetaData, priKey ed25519.PrivateKey) (models.SuiTransactionBlockResponse, error) {
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