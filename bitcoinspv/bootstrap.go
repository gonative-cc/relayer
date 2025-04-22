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
	if err := r.waitForBitcoinNode(ctx); err != nil {
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
		r.logger.Error().Msgf("Failed to submit headers: %v", err)
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
				"Error bootstrapping relayer: %v. Attempts: %d, Max attempts: %d",
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

	blockHeight, err := r.getLCLatestBlockHeight(ctx)
	if err != nil {
		return err
	}

	// Here we are ensuring that the relayer after every restart starts
	// submitting headers from the light clients height - confirmationDepth (usually 6).
	baseHeight := blockHeight - r.btcConfirmationDepth + 1

	// NOTE: here we are fetching only headers, if we want to fetch full blocks change the flag to true
	r.logger.Info().Msg("Fetching headers...")
	blocks, err := r.btcClient.GetBTCTailBlocksByHeight(baseHeight, false)
	if err != nil {
		return fmt.Errorf("failed to get headers: %w", err)
	}

	err = r.btcCache.Init(blocks)
	return err
}

// waitForBitcoinNode ensures that the bitcoin node is synchronized by checking
// that its height is equal or more than the Light Client's height.
// This synchronization is required before proceeding with relayer operations.
func (r *Relayer) waitForBitcoinNode(ctx context.Context) error {
	btcLatestBlockHeight, err := r.getBTCLatestBlockHeight()
	if err != nil {
		return err
	}

	lcLatestBlockHeight, err := r.getLCLatestBlockHeight(ctx)
	if err != nil {
		return err
	}

	if btcLatestBlockHeight == 0 || btcLatestBlockHeight < lcLatestBlockHeight {
		return r.waitForBTCCatchup(ctx, btcLatestBlockHeight, lcLatestBlockHeight)
	}

	return nil
}

func (r *Relayer) getBTCLatestBlockHeight() (int64, error) {
	_, btcLatestBlockHeight, err := r.btcClient.GetBTCTipBlock()
	if err != nil {
		return 0, err
	}

	r.logger.Info().Msgf(
		"BTC latest block height: (%d)",
		btcLatestBlockHeight,
	)

	return btcLatestBlockHeight, nil
}

func (r *Relayer) getLCLatestBlockHeight(ctx context.Context) (int64, error) {
	block, err := r.lcClient.GetLatestBlockInfo(ctx)
	if err != nil {
		return 0, err
	}

	r.logger.Info().Msgf(
		"Light client header chain latest block height: (%d)",
		block.Height,
	)

	return block.Height, nil
}

func (r *Relayer) waitForBTCCatchup(ctx context.Context, btcHeight int64, lcHeight int64) error {
	r.logger.Info().Msgf(
		"BTC chain (length %d) falls behind light client header chain (length %d), wait until BTC catches up",
		btcHeight, lcHeight,
	)

	for {
		// TODO: support ctx cancellation

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
				"BTC (height %d) has synchronized with light client header (height %d), proceeding with bootstrap",
				btcLatestBlockHeight, lcLatestBlockHeight,
			)
			return nil
		}

		r.logger.Info().Msgf(
			"BTC (height %d) is not yet synchronized with light client header (height %d), continuing to wait",
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
