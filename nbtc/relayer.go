package nbtc

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// Relayer broadcasts signed transactions from the database to the Bitcoin network.
type Relayer struct {
	db                      *dal.DB
	btcClient               bitcoin.Client
	shutdownChan            chan struct{}
	processTxsTicker        *time.Ticker
	confirmTxsTicker        *time.Ticker
	txConfirmationThreshold uint8
}

// RelayerConfig holds the configuration parameters for the Relayer.
type RelayerConfig struct {
	ProcessTxsInterval    time.Duration `json:"processTxsInterval"`
	ConfirmTxsInterval    time.Duration `json:"confirmTxsInterval"`
	ConfirmationThreshold uint8         `json:"confirmationThreshold"`
}

// NewRelayer creates a new Relayer instance with the given configuration.
func NewRelayer(
	btcClientConfig rpcclient.ConnConfig,
	relayerConfig RelayerConfig,
	db *dal.DB,
) (*Relayer, error) {

	if db == nil {
		err := fmt.Errorf("database cannot be nil")
		log.Err(err).Msg("")
		return nil, err
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

	if relayerConfig.ProcessTxsInterval == 0 {
		relayerConfig.ProcessTxsInterval = time.Second * 5
	}

	if relayerConfig.ConfirmTxsInterval == 0 {
		relayerConfig.ConfirmTxsInterval = time.Second * 7
	}

	if relayerConfig.ConfirmationThreshold == 0 {
		relayerConfig.ConfirmationThreshold = 6
	}

	return &Relayer{
		db:                      db,
		btcClient:               client,
		shutdownChan:            make(chan struct{}),
		processTxsTicker:        time.NewTicker(relayerConfig.ProcessTxsInterval),
		confirmTxsTicker:        time.NewTicker(relayerConfig.ConfirmTxsInterval),
		txConfirmationThreshold: relayerConfig.ConfirmationThreshold,
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
			if err := r.processSignedTxs(); err != nil {
				var sqliteErr *sqlite3.Error
				//TODO: decide on which exact errors to continue and on which to stop the relayer
				if errors.As(err, &sqliteErr) {
					log.Err(err).Msg("Critical error updating transaction status, shutting down")
					close(r.shutdownChan)
				} else {
					log.Err(err).Msg("Error processing transactions, continuing")
				}
			}
		//TODO: do we need a subroutine for it? Also i think there  might be a race condition on the database
		// so probably we should wrap the db in a mutex
		case <-r.confirmTxsTicker.C:
			if err := r.checkConfirmations(); err != nil {
				var sqliteErr *sqlite3.Error
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

// processSignedTxs processes signed transactions from the database.
func (r *Relayer) processSignedTxs() error {
	signedTxs, err := r.db.GetSignedIkaSignRequests()
	if err != nil {
		return err
	}

	for _, tx := range signedTxs {
		rawTx := make([]byte, 0, len(tx.Payload)+len(tx.FinalSig))
		rawTx = append(rawTx, tx.Payload...)
		rawTx = append(rawTx, tx.FinalSig...)
		var msgTx wire.MsgTx
		if err := msgTx.Deserialize(bytes.NewReader(rawTx)); err != nil {
			return err
		}

		txHash, err := r.btcClient.SendRawTransaction(&msgTx, false)
		if err != nil {
			return fmt.Errorf("error broadcasting transaction: %w", err)
		}
		// TODO: add failed broadcasting to the bitcoinTx table with notes about the error

		err = r.db.InsertBtcTx(dal.BitcoinTx{
			TxID:      tx.ID,
			Status:    dal.Broadcasted,
			BtcTxID:   txHash.CloneBytes(),
			Timestamp: time.Now().Unix(),
			Note:      "",
		})
		if err != nil {
			return fmt.Errorf("DB: can't update tx status: %w", err)
		}

		log.Info().Str("txHash", txHash.String()).Msg("Broadcasted transaction: ")
	}
	return nil
}

// checkConfirmations checks all the broadcasted transactions to bitcoin
// and if confirmed updates the database accordingly.
func (r *Relayer) checkConfirmations() error {
	broadcastedTxs, err := r.db.GetBroadcastedBitcoinTxsInfo()
	if err != nil {
		return err
	}

	for _, tx := range broadcastedTxs {
		hash, err := chainhash.NewHash(tx.BtcTxID)
		if err != nil {
			return err
		}
		txDetails, err := r.btcClient.GetTransaction(hash)
		if err != nil {
			return fmt.Errorf("error getting transaction details: %w", err)
		}

		// TODO: decide what threshold to use. Read that 6 is used on most of the cex'es etc.
		if txDetails.Confirmations >= int64(r.txConfirmationThreshold) {
			err = r.db.UpdateBitcoinTxToConfirmed(tx.TxID, tx.BtcTxID)
			if err != nil {
				return fmt.Errorf("DB: can't update tx status: %w", err)
			}
			log.Info().Msgf("Transaction confirmed: %s", tx.BtcTxID)
		}
	}
	return nil
}

// Stop initiates a shutdown of the relayer.
func (r *Relayer) Stop() {
	close(r.shutdownChan)
}
