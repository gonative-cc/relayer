package suigoclient

import (
	"context"
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/fardream/go-bcs/bcs"
	"github.com/pattonkan/sui-go/sui"
	"github.com/pattonkan/sui-go/sui/suiptb"
	"github.com/pattonkan/sui-go/suiclient"
	"github.com/pattonkan/sui-go/suisigner"
	"github.com/rs/zerolog"
)

type bcsEncode []byte

func getBCSResult(res *suiclient.DevInspectTransactionBlockResponse) ([]bcsEncode, error) {
	bcsEncode := make([]bcsEncode, len(res.Results[0].ReturnValues))

	for i, item := range res.Results[0].ReturnValues {
		var b []byte
		// TODO: Breakdown to simple term
		c := item.([]interface{})[0].([]interface{})
		b = make([]byte, len(c))

		for i, v := range c {
			b[i] = byte(v.(float64))
		}
		bcsEncode[i] = b
	}
	return bcsEncode, nil
}

// LCClient is Bitcoin Light Client
type LCClient struct {
	*suiclient.ClientImpl
	*suisigner.Signer
	logger      zerolog.Logger
	lcPackageID *sui.PackageId
	lcObject    *suiclient.SuiObjectData
}

// ContainsBlock check block hash exist in light client
func (c *LCClient) ContainsBlock(_ context.Context, blockHash chainhash.Hash) (bool, error) {
	ptb := suiptb.NewTransactionDataTransactionBuilder()
	lcObj, err := ptb.Obj(
		suiptb.ObjectArg{
			SharedObject: &suiptb.SharedObjectArg{
				Id:                   c.lcObject.ObjectId,
				InitialSharedVersion: *c.lcObject.Owner.Shared.InitialSharedVersion,
				Mutable:              false,
			},
		},
	)
	if err != nil {
		return false, err
	}

	b, err := ptb.Pure(blockHash[:])
	if err != nil {
		return false, err
	}
	ptb.Command(suiptb.Command{
		MoveCall: &suiptb.ProgrammableMoveCall{
			Package:       c.lcPackageID,
			Module:        "light_client",
			Function:      "exist",
			TypeArguments: []sui.TypeTag{},
			Arguments:     []suiptb.Argument{lcObj, b},
		},
	})

	tx := suiptb.NewTransactionData(c.Address, ptb.Finish(), nil, suiclient.DefaultGasBudget, suiclient.DefaultGasPrice)

	txBytes, err := bcs.Marshal(tx.V1.Kind)

	if err != nil {
		return false, err
	}

	r := suiclient.DevInspectTransactionBlockRequest{
		SenderAddress: c.Address,
		TxKindBytes:   txBytes,
	}

	resp, err := c.DevInspectTransactionBlock(context.Background(), &r)

	if resp.Error != "" {
		return false, errors.New(resp.Error)
	}

	resultEncoded, err := getBCSResult(resp)

	if err != nil {
		return false, err
	}

	var result bool

	err = bcs.UnmarshalAll(resultEncoded[0], &result)

	if err != nil {
		return false, err
	}

	return result, nil
}
