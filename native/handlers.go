package native

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
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
	

	signerAccount, err := CreateSigner(os.Getenv("SIGNER_ACCOUNT_MNEMONIC"))
	if err != nil {
		i.logger.Err(err).Msg("Error creating signer:")
		return err
	}

	i.logger.Info().Int64("light block", lb.SignedHeader.Header.Height).Msg("Light Block ")

	i.logger.Info().Str("signer address", signerAccount.Address).Msg("Light Block ")


	gasObj := os.Getenv("GAS_ADDRESS")

	i.logger.Info().Str("gas address", gasObj).Msg("Light Block ")

	rsp, err := callMoveFunction(ctx, i.cli, signerAccount.Address, gasObj, lb)
	if err != nil {
		i.logger.Err(err).Msg("Error calling move function:")
		return err
	}
	// fmt.Println("Move call response:", rsp)

	rsp2, err := executeTransaction(ctx, i.cli, rsp, signerAccount.PriKey)
	if err != nil {
		i.logger.Err(err).Msg("Error executing transaction:")
		return err
	}
	i.logger.Debug().Any("transaction response", rsp2).Msg("After making trasaction")
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


