package nbtc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/rs/zerolog/log"
)

// Relayer broadcasts pending transactions from the database to the Bitcoin network.
type Relayer struct {
	db           *dal.DB
	btcClient    bitcoin.Client
	shutdownChan chan struct{}
}

// NewRelayer creates a new Relayer instance with the given configuration.
func NewRelayer(db *dal.DB) (*Relayer, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	host := os.Getenv("BTC_RPC")
	user := os.Getenv("BTC_RPC_USER")
	pass := os.Getenv("BTC_RPC_PASS")

	if host == "" || user == "" || pass == "" {
		err := fmt.Errorf("missing env variables with Bitcoin node configuration")
		log.Err(err).Msg("")
		return nil, err
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   false,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}

	return &Relayer{
		db:           db,
		btcClient:    client,
		shutdownChan: make(chan struct{}),
	}, nil
}

// Start starts the relayer's main loop to broadcast transactions.
func (r *Relayer) Start() error {
	ticker := time.NewTicker(time.Second * 2)

	for {
		select {
		case <-r.shutdownChan:
			r.btcClient.Shutdown()
			return nil
		case <-ticker.C:
			if err := r.processPendingTxs(); err != nil {
				//TODO: add proper error handling here and decide which errors we shutdown the relayer
				if err.Error() == "error updating transaction status" {
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
		log.Err(err).Msg("Error getting pending transactions")
		return err
	}

	for _, tx := range pendingTxs {
		decodedTx, err := decodeRawTx(tx.RawTx)
		if err != nil {
			log.Err(err).Msg("Error decoding transaction")
			return err
		}

		txHash, err := r.btcClient.SendRawTransaction(decodedTx, false)
		if err != nil {
			return fmt.Errorf("Error broadcasting transaction: %w", err)
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

// decodeRawTx decodes a raw transaction from a hex string.
func decodeRawTx(rawTxHex string) (*wire.MsgTx, error) {
	serializedTx, err := hex.DecodeString(rawTxHex)
	if err != nil {
		return nil, err
	}

	var msgTx wire.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, err
	}

	return &msgTx, nil
}
