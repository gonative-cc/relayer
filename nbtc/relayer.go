package nbtc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gonative-cc/relayer/dal"
	err "github.com/gonative-cc/relayer/errors"
	"github.com/gonative-cc/relayer/ika2btc"
	"github.com/gonative-cc/relayer/native2ika"
	"github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// Relayer handles the flow of transactions from the Native chain to Bitcoin.
// It uses two processors:
// - nativeProcessor: To send transactions from Native to IKA for signing.
// - btcProcessor: To broadcast signed transactions to Bitcoin and monitor confirmations.
type Relayer struct {
	db               *dal.DB
	nativeProcessor  *native2ika.Processor
	btcProcessor     *ika2btc.Processor
	shutdownChan     chan struct{}
	processTxsTicker *time.Ticker
	confirmTxsTicker *time.Ticker
	// native Sign Request
	signReqTicker     *time.Ticker
	signReqFetcher    native2ika.SignReqFetcher
	signReqStart      int
	signReqFetchLimit int
}

// RelayerConfig holds the configuration parameters for the Relayer.
type RelayerConfig struct {
	ProcessTxsInterval    time.Duration `json:"processTxsInterval"`
	ConfirmTxsInterval    time.Duration `json:"confirmTxsInterval"`
	FetchSignReqInterval  time.Duration `json:"fetchSignReqInterval"`
	ConfirmationThreshold uint8         `json:"confirmationThreshold"`
	FetchFrom             int           `json:"fetchFrom"`
	FetchLimit            int           `json:"fetchLimit"`
}

// NewRelayer creates a new Relayer instance with the given configuration and processors.
func NewRelayer(
	relayerConfig RelayerConfig,
	db *dal.DB,
	nativeProcessor *native2ika.Processor,
	btcProcessor *ika2btc.Processor,
	fetcher native2ika.SignReqFetcher,
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

	if fetcher == nil {
		err := err.ErrNoFetcher
		log.Err(err).Msg("")
		return nil, err
	}

	if relayerConfig.ProcessTxsInterval == 0 {
		relayerConfig.ProcessTxsInterval = time.Second * 5
	}

	if relayerConfig.ConfirmTxsInterval == 0 {
		relayerConfig.ConfirmTxsInterval = time.Second * 7
	}

	if relayerConfig.FetchSignReqInterval == 0 {
		relayerConfig.FetchSignReqInterval = time.Second * 10
	}

	if relayerConfig.ConfirmationThreshold == 0 {
		relayerConfig.ConfirmationThreshold = 6
	}

	return &Relayer{
		db:                db,
		nativeProcessor:   nativeProcessor,
		btcProcessor:      btcProcessor,
		shutdownChan:      make(chan struct{}),
		processTxsTicker:  time.NewTicker(relayerConfig.ProcessTxsInterval),
		confirmTxsTicker:  time.NewTicker(relayerConfig.ConfirmTxsInterval),
		signReqTicker:     time.NewTicker(relayerConfig.FetchSignReqInterval),
		signReqFetcher:    fetcher,
		signReqStart:      relayerConfig.FetchFrom,
		signReqFetchLimit: relayerConfig.FetchLimit,
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
		case <-r.signReqTicker.C:
			if err := r.fetchAndStoreNativeSignRequests(); err != nil {
				var sqliteErr *sqlite3.Error
				if errors.As(err, &sqliteErr) {
					log.Err(err).Msg("Critical error saving signing request, shutting down")
					close(r.shutdownChan)
				} else {
					log.Err(err).Msg("Error processing blocks, continuing")
				}
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

// fetchAndStoreSignRequests fetches and stores sign requests from the Native chain.
func (r *Relayer) fetchAndStoreNativeSignRequests() error {
	signRequests, err := r.signReqFetcher.GetBtcSignRequests(r.signReqStart, r.signReqFetchLimit)
	if err != nil {
		log.Err(err).Msg("Error fetching sign requests from native, continuing")
	}

	for _, sr := range signRequests {
		err = r.storeSignRequest(sr)
		if err != nil {
			return err
		}
	}

	r.signReqStart += 5
	return nil
}

// storeSignRequest stores a single SignReq from the Native chain.
func (r *Relayer) storeSignRequest(signRequest native2ika.SignReq) error {
	err := r.db.InsertIkaSignRequest(dal.IkaSignRequest(signRequest))
	if err != nil {
		return fmt.Errorf("failed to insert IkaSignRequest: %w", err)
	}

	return nil
}

// Stop initiates a shutdown of the relayer.
func (r *Relayer) Stop() {
	close(r.shutdownChan)
}
