package native

import (
	"context"
	"crypto/ed25519"
	"os"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	tmtypes "github.com/cometbft/cometbft/types"
)

func callMoveFunction(ctx context.Context, cli *sui.Client, signerAddress string, gasObj string,
	lb *tmtypes.LightBlock) (models.TxnMetaData, error) {
	return cli.MoveCall(ctx, models.MoveCallRequest{
		Signer:          signerAddress,
		PackageObjectId: os.Getenv("SMART_CONTRACT_ADDRESS"),
		Module:          os.Getenv("SMART_CONTRACT_MODULE"),
		Function:        os.Getenv("SMART_CONTRACT_FUNCTION"),
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			lb,
		},
		Gas:       gasObj,
		GasBudget: os.Getenv("GAS_BUDGET"),
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
