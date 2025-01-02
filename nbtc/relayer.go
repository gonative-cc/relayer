package nbtc

import (
	"context"
	"errors"
	"fmt"
	"time"

	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/gonative-cc/relayer/dal"
	err "github.com/gonative-cc/relayer/errors"
	"github.com/gonative-cc/relayer/ika2btc"
	"github.com/gonative-cc/relayer/native"
	"github.com/gonative-cc/relayer/native2ika"
	"github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// Relayer handles the flow of transactions from the Native chain to Bitcoin.
// It uses two processors:
// - nativeProcessor: To send transactions from Native to IKA for signing.
// - btcProcessor: To broadcast signed transactions to Bitcoin and monitor confirmations.
type Relayer struct {
	db                 *dal.DB
	nativeProcessor    *native2ika.Processor
	btcProcessor       *ika2btc.Processor
	shutdownChan       chan struct{}
	processTxsTicker   *time.Ticker
	confirmTxsTicker   *time.Ticker
	fetchBlocksTicker  *time.Ticker
	blockchain         native.Blockchain
	fetchedBlockHeight int64
}

// RelayerConfig holds the configuration parameters for the Relayer.
type RelayerConfig struct {
	ProcessTxsInterval    time.Duration `json:"processTxsInterval"`
	ConfirmTxsInterval    time.Duration `json:"confirmTxsInterval"`
	FetchBlocksInterval   time.Duration `json:"fetchBlocksInterval"`
	ConfirmationThreshold uint8         `json:"confirmationThreshold"`
	BlockHeigh            int64         `jsoin:"blockHeight"`
}

// NewRelayer creates a new Relayer instance with the given configuration and processors.
func NewRelayer(
	relayerConfig RelayerConfig,
	db *dal.DB,
	nativeProcessor *native2ika.Processor,
	btcProcessor *ika2btc.Processor,
	blockchain native.Blockchain,
) (*Relayer, error) {

	if db == nil {
		err := err.ErrNoDB
		log.Err(err).Msg("")
		return nil, err
	}

	if nativeProcessor == nil {
		err := err.ErrNoNativeProcessor
		log.Err(err).Msg("")
		return nil, err
	}

	if btcProcessor == nil {
		err := err.ErrNoBtcProcessor
		log.Err(err).Msg("")
		return nil, err
	}

	if blockchain == nil {
		err := err.ErrNoBlockchain
		log.Err(err).Msg("")
		return nil, err
	}

	if relayerConfig.ProcessTxsInterval == 0 {
		relayerConfig.ProcessTxsInterval = time.Second * 5
	}

	if relayerConfig.ConfirmTxsInterval == 0 {
		relayerConfig.ConfirmTxsInterval = time.Second * 7
	}

	if relayerConfig.FetchBlocksInterval == 0 {
		relayerConfig.FetchBlocksInterval = time.Second * 10
	}

	if relayerConfig.ConfirmationThreshold == 0 {
		relayerConfig.ConfirmationThreshold = 6
	}

	return &Relayer{
		db:                 db,
		nativeProcessor:    nativeProcessor,
		btcProcessor:       btcProcessor,
		shutdownChan:       make(chan struct{}),
		processTxsTicker:   time.NewTicker(relayerConfig.ProcessTxsInterval),
		confirmTxsTicker:   time.NewTicker(relayerConfig.ConfirmTxsInterval),
		fetchBlocksTicker:  time.NewTicker(relayerConfig.FetchBlocksInterval),
		blockchain:         blockchain,
		fetchedBlockHeight: relayerConfig.BlockHeigh,
	}, nil
}

// Start starts the relayer's main loop.
func (r *Relayer) Start(ctx context.Context) error {
	for {
		select {
		case <-r.shutdownChan:
			r.btcProcessor.Shutdown()
			return nil
		case <-r.processTxsTicker.C:
			_ = r.processSignRequests(ctx)
			_ = r.processSignedTxs()
		//TODO: do we need a subroutine for it? Also i think there  might be a race condition on the database
		// so probably we should wrap the db in a mutex
		case <-r.confirmTxsTicker.C:
			if err := r.btcProcessor.CheckConfirmations(); err != nil {
				var sqliteErr *sqlite3.Error
				if errors.As(err, &sqliteErr) {
					log.Err(err).Msg("Critical error updating transaction status, shutting down")
					close(r.shutdownChan)
				} else {
					log.Err(err).Msg("Error processing transactions, continuing")
				}
			}
		case <-r.fetchBlocksTicker.C:
			if err := r.fetchAndProcessNativeBlocks(ctx); err != nil {
				//TODO: add error handling here
			}
		}
	}
}

// processSignRequests processes signd requests from the Native chain.
func (r *Relayer) processSignRequests(ctx context.Context) error {
	if err := r.nativeProcessor.Run(ctx); err != nil {
		var sqliteErr *sqlite3.Error
		//TODO: decide on which exact errors to continue and on which to stop the relayer
		if errors.As(err, &sqliteErr) {
			log.Err(err).Msg("Critical error with the database, shutting down")
			close(r.shutdownChan)
		} else {
			log.Err(err).Msg("Error processing native transactions, continuing")
		}
		return err
	}
	return nil
}

// processSignedTxs processes signed transactions and broadcasts them to Bitcoin.
func (r *Relayer) processSignedTxs() error {
	if err := r.btcProcessor.Run(); err != nil {
		var sqliteErr *sqlite3.Error
		//TODO: decide on which exact errors to continue and on which to stop the relayer
		if errors.As(err, &sqliteErr) {
			log.Err(err).Msg("Critical error updating transaction status, shutting down")
			close(r.shutdownChan)
		} else {
			log.Err(err).Msg("Error processing bitcoin transactions, continuing")
		}
		return err
	}
	return nil
}

// fetchAndProcessNativeBlocks fetches and processes blocks from the Native chain.
func (r *Relayer) fetchAndProcessNativeBlocks(ctx context.Context) error {
	//TODO: decide how many blocks we want to process at a time, do we process all of them, or a limited amount of blocks??
	startHeight := r.fetchedBlockHeight
	for height := startHeight; height < startHeight+20; height++ {
		block, _, err := r.blockchain.Block(ctx, height)
		if err != nil {
			log.Err(err).Msg("Error fetching block, continuing")
			continue
		}

		err = r.processNativeBlock(block)
		if err != nil {
			log.Err(err).Msg("Error processing block, continuing")
			continue
		}

		r.fetchedBlockHeight = height
	}
	return nil
}

// processNativeBlock processes a single block from the Native chain.
func (r *Relayer) processNativeBlock(block *tmtypes.Block) error {
	//TODO: add proper information extraction logic from the block events here
	signRequest := dal.IkaSignRequest{
		ID:        uint64(block.Height),
		Payload:   []byte("payload"),
		DWalletID: "dwallet_id",
		UserSig:   "user_sig",
		FinalSig:  nil,
		Timestamp: time.Now().Unix(),
	}

	err := r.db.InsertIkaSignRequest(signRequest)
	if err != nil {
		return fmt.Errorf("failed to insert IkaSignRequest: %w", err)
	}

	return nil
}

// Stop initiates a shutdown of the relayer.
func (r *Relayer) Stop() {
	close(r.shutdownChan)
}
