package native

import (
	"context"
	"encoding/hex"
	"fmt"
	"reflect"
	tmtypes "github.com/cometbft/cometbft/types"
)


// HandleNewBlock handles the receive of new block from the chain.
func (i *Indexer) HandleNewBlock(ctx context.Context, blk *tmtypes.Block) error {
	// todo set current block

	// and continues to handle a block normally.
	return i.HandleBlock(ctx, blk)
}

// HandleBlock handles the receive of an block from the chain.
func (i *Indexer) HandleBlock(ctx context.Context, blk *tmtypes.Block) error {
	// light block
	lb, err:= i.b.LightProvider().LightBlock(ctx, blk.Header.Height)
	if err != nil {
		return err
	}
	i.logger.Info().Int64("light block", lb.SignedHeader.Header.Height).Msg("Light Block ")

	fmt.Println(reflect.TypeOf(lb))

	signerAccount, err := CreateSigner("defense frost latin party smoke veteran bamboo dignity sniff eyebrow extra lottery")
	if err != nil {
		fmt.Println("Error creating signer:", err)
		return err
	}
	fmt.Println(signerAccount.Address)

	gasObj := "0x12c7c74c3c7892156fa3429e9157992363fe68bf8f7d58a7d781d09ea4a74802"

	rsp, err := callMoveFunction(ctx, i.cli, signerAccount.Address, gasObj, lb)
	if err != nil {
		fmt.Println("Error calling move function:", err)
		return err
	}
	//fmt.Println("Move call response:", rsp)

	rsp2, err := executeTransaction(ctx, i.cli, rsp, signerAccount.PriKey)
	if err != nil {
		fmt.Println("Error executing transaction:", err)
		return err
	}

	fmt.Println(rsp2)



	for _, tx := range blk.Data.Txs {
		if err := i.HandleTx(ctx, int(blk.Header.Height), int(blk.Time.Unix()), tx); err != nil {
			i.logger.Err(err).Int64("height", blk.Height).Msg("error handling block")
			continue
		}
	}
	return nil
}

// HandleTx handles the receive of new Tx from the chain.
//
//revive:disable:unused-parameter
func (i *Indexer) HandleTx(ctx context.Context, blockHeight, blockTimeUnix int, tmTx tmtypes.Tx) error {
	tx, err := i.b.DecodeTx(tmTx)
	if err != nil {
		i.logger.Err(err).Msg("error decoding Tx")
		return err
	}

	txHash := tmTx.Hash()
	i.logger.Debug().Msg("handling tx: " + hex.EncodeToString(txHash))

	for _, msg := range tx.GetMsgs() {
		i.logger.Debug().Any("msg", msg).Msg("handling msg")
	}
	return nil
}


