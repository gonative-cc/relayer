package bitcoinspv

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
	relayertypes "github.com/gonative-cc/relayer/bitcoinspv/types"
)

var (
	bootstrapRetryAttempts = uint(60)
	bootstrapRetryInterval = retry.Delay(30 * time.Second)
)

func (r *Relayer) bootstrapRelayer(ctx context.Context, skipSubscription bool) error {
	if err := r.waitForBitcoinCatchup(ctx); err != nil {
		return err
	}

	if err := r.setupCache(ctx, skipSubscription); err != nil {
		return err
	}

	if err := r.processAndTrimCache(ctx); err != nil {
		return err
	}

	r.logger.Info().Msgf("BTC cache size: %d", r.btcCache.Size())
	r.logger.Info().Msg("Successfully bootstrapped")
	return nil
}

func (r *Relayer) setupCache(ctx context.Context, skipSubscription bool) error {
	if err := r.initializeBTCCache(ctx); err != nil {
		return err
	}

	if !skipSubscription {
		r.btcClient.SubscribeNewBlocks()
	}
	return nil
}

func (r *Relayer) processAndTrimCache(ctx context.Context) error {
	if err := r.processHeaders(ctx); err != nil {
		return err
	}

	err := r.resizeAndTrimCache()
	return err
}

func (r *Relayer) processHeaders(ctx context.Context) error {
	headersToProcess := r.btcCache.GetAllBlocks()
	if _, err := r.ProcessHeaders(ctx, headersToProcess); err != nil {
		// occurs when multiple competing spv relayers exist
		// or when our btc node is not fully synchronized
		r.logger.Err(err).Msg("Failed to submit headers")
		return err
	}
	return nil
}

func (r *Relayer) resizeAndTrimCache() error {
	if err := r.btcCache.Resize(r.btcConfirmationDepth); err != nil {
		return err
	}
	r.btcCache.Trim()
	return nil
}

func (r *Relayer) createRelayerContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	r.wg.Add(1)

	go func() {
		defer func() {
			cancel()
			r.wg.Done()
		}()

		select {
		case <-r.quitChan():
		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}

func (r *Relayer) multitryBootstrap(skipSubscription bool) {
	ctx, cancel := r.createRelayerContext()
	defer cancel()

	retryOpts := r.getBootstrapRetryOptions(ctx)

	if err := retry.Do(
		func() error {
			return r.bootstrapRelayer(ctx, skipSubscription)
		},
		retryOpts...,
	); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		r.logger.Fatal().Msgf("Failed to bootstrap relayer: %v after %d attempts", err, bootstrapRetryAttempts)
	}
}

func (r *Relayer) getBootstrapRetryOptions(ctx context.Context) []retry.Option {
	return []retry.Option{
		retry.Context(ctx),
		retry.Attempts(bootstrapRetryAttempts),
		bootstrapRetryInterval,
		retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			r.logger.Warn().Msgf(
				"Failed bootstrapping relayer: %v. Attempts: %d, Max attempts: %d",
				err, n+1, bootstrapRetryAttempts,
			)
		}),
	}
}

// initializeBTCCache initializes the BTC cache with blocks from T-k to T
// where T is the height of the latest block in the light client
// and k is the confirmation depth
func (r *Relayer) initializeBTCCache(ctx context.Context) error {
	cache, err := relayertypes.NewBTCCache(r.Config.BTCCacheSize)
	if err != nil {
		return err
	}
	r.btcCache = cache

	lcHeight, err := r.getLCLatestBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch light client height: %w", err)
	}

	indexerHeight, err := r.getIndexerLatestBlockHeight()
	if err != nil {
		return fmt.Errorf("failed to fetch indexer height: %w", err)
	}

	_, btcTipHeight, err := r.btcClient.GetBTCTipBlock()
	if err != nil {
		return fmt.Errorf("failedt bitcoin node tip height: %w", err)
	}

	r.logger.Info().
		Int64("light_client", lcHeight).
		Int64("indexer", indexerHeight).
		Int64("btc_node", btcTipHeight).
		Msg("Latest known heights")

	if r.btcIndexer != nil && indexerHeight < btcTipHeight {
		r.logger.Info().Int64("from", indexerHeight+1).Int64("to", btcTipHeight).Msg("Starting indexer backfill...")
		if err := r.backfillIndexer(ctx, indexerHeight+1, btcTipHeight); err != nil {
			return fmt.Errorf("indexer backfill failed: %w", err)
		}
	}

	// Here we are ensuring that the relayer after every restart starts
	// submitting headers from the light clients height - confirmationDepth (usually 6).
	startSyncHeight := lcHeight - r.btcConfirmationDepth + 1

	r.logger.Info().Int64("start_height", startSyncHeight).Msgf("Bootstrapping lc with headers")

	headers, err := r.btcClient.GetBTCTailBlocksByHeight(startSyncHeight, false) // false - headers only
	if err != nil {
		return fmt.Errorf("failed to get headers for bootstrap: %w", err)
	}
	return r.btcCache.Init(headers)
}

func (r *Relayer) getBTCLatestBlockHeight() (int64, error) {
	_, btcLatestBlockHeight, err := r.btcClient.GetBTCTipBlock()
	if err != nil {
		return 0, err
	}

	return btcLatestBlockHeight, nil
}

func (r *Relayer) getLCLatestBlockHeight(ctx context.Context) (int64, error) {
	block, err := r.lcClient.GetLatestBlockInfo(ctx)
	if err != nil {
		return 0, err
	}

	return block.Height, nil
}

// waitForBitcoinCatchup ensures that the bitcoin node is synchronized by checking
// that its height is equal or more than the Light Client's height.
// This synchronization is required before proceeding with relayer operations.
func (r *Relayer) waitForBitcoinCatchup(ctx context.Context) error {
	firstRun := true
	for {
		btcLatestBlockHeight, err := r.getBTCLatestBlockHeight()
		if err != nil {
			return err
		}

		lcLatestBlockHeight, err := r.getLCLatestBlockHeight(ctx)
		if err != nil {
			return err
		}

		if isBTCCaughtUp(btcLatestBlockHeight, lcLatestBlockHeight) {
			r.logger.Info().Msgf(
				"BTC (height %d) has synchronized, latest block matches the light client header (height %d)",
				btcLatestBlockHeight, lcLatestBlockHeight,
			)
			return nil
		}

		logger := r.logger.Debug()
		if firstRun {
			logger = r.logger.Info()
			firstRun = false
		}
		logger.Msgf(
			"BTC chain (length %d) falls behind light client header chain (length %d). Waiting until BTC catches up...",
			btcLatestBlockHeight, lcLatestBlockHeight,
		)

		select {
		case <-time.After(r.catchupLoopWait):
			// Continue the loop after the timeout
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func isBTCCaughtUp(btcHeight int64, lcHeight int64) bool {
	return btcHeight > 0 && btcHeight >= lcHeight
}

func (r *Relayer) backfillIndexer(ctx context.Context, startHeight, endHeight int64) error {
	batchSize := int64(20)

	for current := startHeight; current <= endHeight; current += batchSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		batchEnd := current + batchSize - 1
		if batchEnd > endHeight {
			batchEnd = endHeight
		}

		r.logger.Info().Int64("from", current).Int64("to", batchEnd).Msg("Fetching block batch for indexer backfill")

		blocksInBatch := make([]*relayertypes.IndexedBlock, 0, batchEnd-current+1)
		for i := current; i <= batchEnd; i++ {
			block, err := r.btcClient.GetBTCBlockByHeight(i)
			if err != nil {
				return fmt.Errorf("failed to fetch block at height %d for indexer backfill: %w", i, err)
			}
			blocksInBatch = append(blocksInBatch, block)
		}

		if len(blocksInBatch) > 0 {
			if err := r.btcIndexer.SendBlocks(ctx, blocksInBatch); err != nil {
				return fmt.Errorf("failed to send block batch to indexer: %w", err)
			}
		}
	}
	r.logger.Info().Msg("Indexer backfill completed successfully.")
	return nil
}

func (r *Relayer) getIndexerLatestBlockHeight() (int64, error) {
	if r.btcIndexer == nil {
		return 0, nil
	}
	return := r.btcIndexer.GetLatestHeight()
}
