package nbtc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"errors"
	"fmt"
	"time"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// Relayer broadcasts pending transactions from the database to the Bitcoin network.
type Relayer struct {
	db           *dal.DB
	btcClient    bitcoin.Client
	shutdownChan chan struct{}
  processTxsTicker *time.Ticker
}

// NewRelayer creates a new Relayer instance with the given configuration.
func NewRelayer(btcClientConfig rpcclient.ConnConfig, processTxsInterval time.Duration, db *dal.DB) (*Relayer, error) {

	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	if btcClientConfig.Host == "" || btcClientConfig.User == "" || btcClientConfig.Pass == "" {
		err := fmt.Errorf("missing bitcoin node configuration")
		log.Err(err).Msg("")
		return nil, err
	}

	client, err := rpcclient.New(&btcClientConfig, nil)
	if err != nil {
		return nil, err
	}

	if processTxsInterval == 0 {
		processTxsInterval = time.Second * 5
	}

	return &Relayer{
		db:               db,
		btcClient:        client,
		shutdownChan:     make(chan struct{}),
		processTxsTicker: time.NewTicker(processTxsInterval),
	}, nil
}

// Start starts the relayer's main loop to broadcast transactions.
func (r *Relayer) Start() error {
	ticker2 := time.NewTicker(time.Second * 4) //TODO: this probably can run every minute or even more

	for {
		select {
		case <-r.shutdownChan:
			r.btcClient.Shutdown()
			return nil
		case <-r.processTxsTicker.C:
			if err := r.processPendingTxs(); err != nil {
				var sqliteErr *sqlite3.Error
				//TODO: decide on which exact errors to continue and on which to stop the relayer
				if errors.As(err, &sqliteErr) {
					log.Err(err).Msg("Critical error updating transaction status, shutting down")
					close(r.shutdownChan)
				} else {
					log.Err(err).Msg("Error processing transactions, continuing")
				}
			}
		case <-ticker2.C:
			r.checkConfirmations()
		}
	}
}

// processPendingTxs processes pending transactions from the database.
func (r *Relayer) processPendingTxs() error {
	pendingTxs, err := r.db.GetPendingTxs()
	if err != nil {
		return err
	}

	for _, tx := range pendingTxs {
		var msgTx wire.MsgTx
		if err := msgTx.Deserialize(bytes.NewReader(tx.RawTx)); err != nil {
			return err
		}

		txHash, err := r.btcClient.SendRawTransaction(&msgTx, false)
		if err != nil {
			return fmt.Errorf("error broadcasting transaction: %w", err)
		}

		err = r.db.UpdateTxStatus(tx.BtcTxID, dal.StatusBroadcasted)
		if err != nil {
			return fmt.Errorf("DB: can't update tx status: %w", err)
		}

		log.Info().Str("txHash", txHash.String()).Msg("Broadcasted transaction: ")
	}
	return nil
}

// checkConfirmations checks all the broadcasted transactions to bitcoin and if confirmed updates the database accordingly.
func (r *Relayer) checkConfirmations() {
	broadcastedTxs, err := r.db.GetBroadcastedTxs()
	if err != nil {
		log.Err(err).Msg("Error getting broadcasted transactions")
		return
	}

	for _, tx := range broadcastedTxs {
		hash, _ := chainhash.NewHash(tx.Hash)
		txDetails, err := r.btcClient.GetTransaction(hash)
		if err != nil {
			log.Err(err).Msgf("Error getting transaction details for txid: %d", tx.BtcTxID)
			return
		}

		if txDetails.Confirmations >= 6 { // TODO: decide what threshold to use. Read that 6 is used on most of the cex'es etc.
			err = r.db.UpdateTxStatus(tx.BtcTxID, dal.StatusConfirmed)
			if err != nil {
				log.Err(err).Msgf("Error updating transaction status for txid: %d", tx.BtcTxID)
			} else {
				log.Info().Msgf("Transaction confirmed: %s", tx.Hash)
			}
		}
	}
}

// Stop initiates a shutdown of the relayer.
func (r *Relayer) Stop() {
	close(r.shutdownChan)
}
