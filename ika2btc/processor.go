package ika2btc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/rs/zerolog/log"
)

// Processor handles processing transactions from IKA to Bitcoin.
type Processor struct {
	BtcClient               bitcoin.Client
	db                      *dal.DB
	txConfirmationThreshold uint8
}

// NewProcessor creates a new Processor instance.
func NewProcessor(
	btcClientConfig rpcclient.ConnConfig,
	confirmationThreshold uint8,
	db *dal.DB,
) (*Processor, error) {

	if db == nil {
		return nil, dal.ErrNoDB
	}

	if btcClientConfig.Host == "" || btcClientConfig.User == "" || btcClientConfig.Pass == "" {
		return nil, bitcoin.ErrNoBtcConfig
	}

	// log.Debug().Msg(fmt.Sprintf("creating a btc client, config: %s, %s, %s", btcClientConfig.User, btcClientConfig.Host, btcClientConfig.Pass))

	client, err := rpcclient.New(&btcClientConfig, nil)
	if err != nil {
		return nil, err
	}

	return &Processor{
		BtcClient:               client,
		db:                      db,
		txConfirmationThreshold: confirmationThreshold,
	}, nil
}

// Run starts a loop to query and process signed transactions from the database.
func (p *Processor) Run() error {
	signedTxs, err := p.db.GetBitcoinTxsToBroadcast()
	if err != nil {
		return err
	}

	log.Info().Msg("\x1b[33mBroadcasting transaction to Bitcoin network...\x1b[0m")

	for _, tx := range signedTxs {
		// rawTx := make([]byte, 0, len(tx.Payload)+len(tx.FinalSig))
		// rawTx = append(rawTx, tx.Payload...)
		// rawTx = append(rawTx, tx.FinalSig...)
		// log.Debug().Msg(fmt.Sprintf("final sig before deserialize: %s", tx.FinalSig))

		// Convert the ASCII representation to raw bytes
		serializedTx, err := hex.DecodeString(string(tx.FinalSig)) // Convert tx.FinalSig to string first
		if err != nil {
			log.Err(err).Msg("hex decode error")
			return err
		}
		var msgTx wire.MsgTx
		if err := msgTx.Deserialize(bytes.NewReader(serializedTx)); err != nil {
			log.Err(err).Msg("deserialize error")
			return err
		}

		// log.Debug().Msg(fmt.Sprintf("serialziedTx: %s", serializedTx))
		// log.Debug().Msg(fmt.Sprintf("tx.FinalSig: %s", tx.FinalSig))

		// for i := 0; i < len(serializedTx); i++ {
		// 	if serializedTx[i] != tx.FinalSig[i] {
		// 		log.Debug().Msg(fmt.Sprintf("Difference at index %d: %x vs %x\n", i, serializedTx[i], tx.FinalSig[i]))
		// 	}
		// }

		// TODO: print here what is happening and find out why we cannot broadcast it

		txHash, err := p.BtcClient.SendRawTransaction(&msgTx, true)
		if err != nil {
			return fmt.Errorf("error broadcasting transaction: %w", err)
		}
		// TODO: add failed broadcasting to the bitcoinTx table with notes about the error

		err = p.db.InsertBtcTx(dal.BitcoinTx{
			SrID:      tx.ID,
			Status:    dal.Broadcasted,
			BtcTxID:   txHash.CloneBytes(),
			Timestamp: time.Now().Unix(),
			Note:      "",
		})
		if err != nil {
			return fmt.Errorf("DB: can't update tx status: {tx: %d, err: %w}", tx.ID, err)
		}
		log.Info().Msgf("\x1b[32mSUCCESS\x1b[0m Broadcasted transaction to Bitcoin: txHash = %s", txHash.String())
	}
	return nil
}

// CheckConfirmations checks all the broadcasted transactions to bitcoin
// and if confirmed updates the database accordingly.
func (p *Processor) CheckConfirmations() error {
	_, err := p.db.GetBroadcastedBitcoinTxsInfo()
	if err != nil {
		return err
	}

	// for _, tx := range broadcastedTxs {
	// 	hash, err := chainhash.NewHash(tx.BtcTxID)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	txDetails, err := p.BtcClient.GetTransaction(hash)
	// 	if err != nil {
	// 		return fmt.Errorf("error getting transaction details: %w", err)
	// 	}

	// 	// TODO: decide what threshold to use. Read that 6 is used on most of the cex'es etc.
	// 	if txDetails.Confirmations >= int64(p.txConfirmationThreshold) {
	// 		err = p.db.UpdateBitcoinTxToConfirmed(tx.TxID, tx.BtcTxID)
	// 		if err != nil {
	// 			return fmt.Errorf("DB: can't update tx status: %w", err)
	// 		}
	// 		log.Info().Msgf("Transaction confirmed: %s", tx.BtcTxID)
	// 	}
	// }
	return nil
}

// Shutdown shuts down the Bitcoin RPC client.
func (p *Processor) Shutdown() {
	p.BtcClient.Shutdown()
}
