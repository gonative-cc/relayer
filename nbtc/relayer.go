package nbtc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/ika2btc"
	"github.com/gonative-cc/relayer/ika2btc/bitcoin"
	"github.com/gonative-cc/relayer/native"
	"github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// Relayer handles the flow of transactions from the Native chain to Bitcoin.
// It uses two processors:
// - nativeProcessor: To send transactions from Native to IKA for signing.
// - btcProcessor: To broadcast signed transactions to Bitcoin and monitor confirmations.
type Relayer struct {
	db               dal.DB
	signReqFetcher   native.SignReqFetcher
	btcProcessor     *ika2btc.Processor
	shutdownChan     chan struct{}
	processTxsTicker *time.Ticker
	confirmTxsTicker *time.Ticker
	signReqTicker    *time.Ticker
	// ID of the first sign req that we want to fetch in the next round
	signReqFetchFrom  int
	signReqFetchLimit int
}

// RelayerConfig holds the configuration parameters for the Relayer.
type RelayerConfig struct {
	ProcessTxsInterval   time.Duration
	ConfirmTxsInterval   time.Duration
	SignReqFetchInterval time.Duration
	// ID of the first sign req that we want to fetch in
	SignReqFetchFrom  int
	SignReqFetchLimit int
}

// These values are used if the corresponding intervals are not provided.
const (
	defaultProcessTxsInterval   = 5 * time.Second
	defaultConfirmTxsInterval   = 7 * time.Second
	defaultSignReqFetchInterval = 10 * time.Second
)

// NewRelayer creates a new Relayer instance with the given configuration and processors.
// TODO: need to add Ika signatures querier
func NewRelayer(
	relayerConfig RelayerConfig,
	db dal.DB,
	btcProcessor *ika2btc.Processor,
	fetcher native.SignReqFetcher,
) (*Relayer, error) {
	if btcProcessor == nil {
		return nil, fmt.Errorf("relayer: %w", bitcoin.ErrNoBtcProcessor)
	}

	if fetcher == nil {
		return nil, fmt.Errorf("relayer: %w", native.ErrNoFetcher)
	}

	if relayerConfig.ProcessTxsInterval == 0 {
		relayerConfig.ProcessTxsInterval = defaultProcessTxsInterval
	}

	if relayerConfig.ConfirmTxsInterval == 0 {
		relayerConfig.ConfirmTxsInterval = defaultConfirmTxsInterval
	}

	if relayerConfig.SignReqFetchInterval == 0 {
		relayerConfig.SignReqFetchInterval = defaultSignReqFetchInterval
	}

	return &Relayer{
		db:                db,
		btcProcessor:      btcProcessor,
		shutdownChan:      make(chan struct{}),
		processTxsTicker:  time.NewTicker(relayerConfig.ProcessTxsInterval),
		confirmTxsTicker:  time.NewTicker(relayerConfig.ConfirmTxsInterval),
		signReqTicker:     time.NewTicker(relayerConfig.SignReqFetchInterval),
		signReqFetcher:    fetcher,
		signReqFetchFrom:  relayerConfig.SignReqFetchFrom,
		signReqFetchLimit: relayerConfig.SignReqFetchLimit,
	}, nil
}

// Start starts the relayer's main loop.
func (r *Relayer) Start(ctx context.Context) error {

	nativeCtx, nativeCancel := context.WithCancel(ctx)
	defer nativeCancel()

	btcCtx, btcCancel := context.WithCancel(ctx)
	defer btcCancel()

	fetchCtx, fetchCancel := context.WithCancel(ctx)
	defer fetchCancel()

	for {
		select {
		case <-r.shutdownChan:
			log.Info().Msg("Relayer stopped successfully")
			return nil
		case <-r.processTxsTicker.C:
			go r.runProcessor(func() error { return r.processSignRequests(nativeCtx) }, "processSignRequests")
			go r.runProcessor(func() error { return r.processSignedTxs(btcCtx) }, "processSignedTxs")
		case <-r.confirmTxsTicker.C:
			go r.runProcessor(func() error { return r.btcProcessor.CheckConfirmations(btcCtx) }, "CheckConfirmations")
		case <-r.signReqTicker.C:
			go r.runProcessor(
				func() error { return r.fetchAndStoreNativeSignRequests(fetchCtx) },
				"fetchAndStoreNativeSignRequests",
			)
		}
	}
}

// handleError handles errors and logs them, potentially shutting down the relayer.
func (r *Relayer) handleError(err error, operation string) {
	var sqliteErr *sqlite3.Error
	//TODO: decide on which exact errors to continue and on which to stop the relayer
	if errors.As(err, &sqliteErr) {
		log.Error().Err(err).Str("operation", operation).Msg("Critical database error, shutting down")
		close(r.shutdownChan)
	} else {
		log.Error().Err(err).Str("operation", operation).Msg("Error in operation , continuing")
	}
}

// processSignRequests processes signd requests from the Native chain.
func (r *Relayer) processSignRequests(_ context.Context) error {
	// TODO: implement this
	panic("not implemented")
}

// processSignedTxs processes signed transactions and broadcasts them to Bitcoin.
func (r *Relayer) processSignedTxs(ctx context.Context) error {
	return r.btcProcessor.Run(ctx)
}

// fetchAndStoreSignRequests fetches and stores sign requests from the Native chain.
func (r *Relayer) fetchAndStoreNativeSignRequests(ctx context.Context) error {
	log.Info().Msg("Fetching sign requests from Native...")

	signRequests, err := r.signReqFetcher.GetBtcSignRequests(r.signReqFetchFrom, r.signReqFetchLimit)
	if err != nil {
		return fmt.Errorf("fetchAndStore: %w", err)
	}
	log.Info().Msgf("SUCCESS: Fetched %d sign requests from Native.", len(signRequests))
	for _, sr := range signRequests {
		err = r.storeSignRequest(ctx, sr)
		if err != nil {
			return fmt.Errorf("fetchAndStore: %w", err)
		}
	}
	r.signReqFetchFrom += r.signReqFetchLimit
	return nil
}

// storeSignRequest stores a single SignReq from the Native chain.
func (r *Relayer) storeSignRequest(ctx context.Context, signRequest native.SignReq) error {
	err := r.db.InsertIkaSignRequest(ctx, dal.IkaSignRequest(signRequest))
	if err != nil {
		return fmt.Errorf("failed to insert IkaSignRequest: %w", err)
	}

	return nil
}

func (r *Relayer) runProcessor(f func() error, name string) {
	if err := f(); err != nil {
		r.handleError(err, name)
	}
}

// Stop initiates a shutdown of the relayer.
func (r *Relayer) Stop() {
	r.processTxsTicker.Stop()
	r.confirmTxsTicker.Stop()
	r.signReqTicker.Stop()
	r.btcProcessor.Shutdown()
	close(r.shutdownChan)
}
