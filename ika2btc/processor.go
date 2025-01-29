package ika2btc

import (
	"bytes"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/rs/zerolog/log"
)

// Processor handles processing transactions from IKA to Bitcoin.
type Processor struct {
	BtcClient               bitcoin.Client
	db                      dal.DB
	txConfirmationThreshold uint8
}

// NewProcessor creates a new Processor instance.
func NewProcessor(
	btcClientConfig rpcclient.ConnConfig,
	confirmationThreshold uint8,
	db dal.DB,
) (*Processor, error) {
	if btcClientConfig.Host == "" || btcClientConfig.User == "" || btcClientConfig.Pass == "" {
		return nil, bitcoin.ErrNoBtcConfig
	}

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
	log.Info().Msg("Broadcasting transaction to Bitcoin network...")
	for _, tx := range signedTxs {
		rawTx := make([]byte, 0, len(tx.Payload)+len(tx.FinalSig))
		rawTx = append(rawTx, tx.Payload...)
		rawTx = append(rawTx, tx.FinalSig...)
		var msgTx wire.MsgTx
		if err := msgTx.Deserialize(bytes.NewReader(rawTx)); err != nil {
			return err
		}

		txHash, err := p.BtcClient.SendRawTransaction(&msgTx, false)
		if err != nil {
			return fmt.Errorf("error broadcasting transaction: %w", err)
		}
		log.Info().Msgf("SUCCESS: Broadcasted transaction to Bitcoin: txHash = %s", txHash.String())
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
	}
	return nil
}

// CheckConfirmations checks all the broadcasted transactions to bitcoin
// and if confirmed updates the database accordingly.
func (p *Processor) CheckConfirmations() error {
	broadcastedTxs, err := p.db.GetBroadcastedBitcoinTxsInfo()
	if err != nil {
		return err
	}

	for _, tx := range broadcastedTxs {
		hash, err := chainhash.NewHash(tx.BtcTxID)
		if err != nil {
			return err
		}
		txDetails, err := p.BtcClient.GetTransaction(hash)
		if err != nil {
			return fmt.Errorf("error getting transaction details: %w", err)
		}

		// TODO: decide what threshold to use. Read that 6 is used on most of the cex'es etc.
		if txDetails.Confirmations >= int64(p.txConfirmationThreshold) {
			err = p.db.UpdateBitcoinTxToConfirmed(tx.TxID, tx.BtcTxID)
			if err != nil {
				return fmt.Errorf("DB: can't update tx status: %w", err)
			}
			log.Info().Msgf("Transaction confirmed: %s", tx.BtcTxID)
		}
	}
	return nil
}

// Shutdown shuts down the Bitcoin RPC client.
func (p *Processor) Shutdown() {
	p.BtcClient.Shutdown()
}
