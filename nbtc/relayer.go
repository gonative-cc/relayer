package nbtc

import (
	"bytes"
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
	db               *dal.DB
	btcClient        bitcoin.Client
	shutdownChan     chan struct{}
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

// Stop initiates a shutdown of the relayer.
func (r *Relayer) Stop() {
	close(r.shutdownChan)
}
