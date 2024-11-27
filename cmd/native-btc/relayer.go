package main

import (
	"bytes"
	"encoding/hex"
	"os"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/rs/zerolog/log"
)

// Config holds the configuration parameters for the Relayer.
type Config struct {
	DatabasePath string `json:"databasePath"`
	BitcoinNode  string `json:"bitcoinNode"`
}

// Relayer broadcasts pending transactions from the database to the Bitcoin network.
type Relayer struct {
	config       Config
	db           *dal.DB
	btcClient    bitcoin.BitcoinClient
	shutdownChan chan struct{}
}

// NewRelayer creates a new Relayer instance with the given configuration.
func NewRelayer(config Config, db *dal.DB) (*Relayer, error) {
	var err error
	if db == nil {
		db, err = dal.NewDB(config.DatabasePath)
		if err != nil {
			return nil, err
		}
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         os.Getenv("BTC_RPC"),
		User:         os.Getenv("BTC_RPC_USER"),
		Pass:         os.Getenv("BTC_RPC_PASS"),
		HTTPPostMode: true,
		DisableTLS:   false,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}

	return &Relayer{
		config:       config,
		db:           db,
		btcClient:    client,
		shutdownChan: make(chan struct{}),
	}, nil
}

// Start starts the relayer's main loop to broadcast transactions.
func (r *Relayer) Start() {
	go func() {
		<-r.shutdownChan
		err := r.db.Close()
		if err != nil {
			log.Err(err).Msg("Error closing database connection")
		}
		r.btcClient.Shutdown()
	}()

	for {
		pendingTxs, err := r.db.GetPendingTxs()
		if err != nil {
			log.Err(err).Msg("Error getting pending transactions")
			continue
		}

		for _, tx := range pendingTxs {
			decodedTx, err := decodeRawTx(tx.RawTx)
			if err != nil {
				log.Err(err).Msg("Error decoding transaction")
				continue
			}

			txHash, err := r.btcClient.SendRawTransaction(decodedTx, false)
			if err != nil {
				log.Err(err).Msg("Error broadcasting transaction")
				continue
			}

			err = r.db.UpdateTxStatus(tx.BtcTxID, dal.StatusBroadcasted)
			if err != nil {
				log.Err(err).Msg("Error updating transaction status")
			}

			log.Info().Str("txHash", txHash.String()).Msg("Broadcasted transaction: ")
		}

		time.Sleep(time.Second * 10)
	}
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
