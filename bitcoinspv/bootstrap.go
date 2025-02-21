package bitcoinspv

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go/v4"
	relayertypes "github.com/gonative-cc/relayer/bitcoinspv/types"
)

var (
	bootstrapRetryAttempts = uint(60)
	bootstrapSyncTicker    = 10 * time.Second
	bootstrapRetryInterval = retry.Delay(30 * time.Second)
)

func (r *Relayer) bootstrapRelayer(skipSubscription bool) error {
	if err := r.initializeAndSync(); err != nil {
		return err
	}

	if err := r.setupCache(skipSubscription); err != nil {
		return err
	}

	if err := r.processAndTrimCache(); err != nil {
		return err
	}

	r.logger.Infof("BTC cache size: %d", r.btcCache.Size())
	r.logger.Info("Successfully bootstrapped")
	return nil
}

func (r *Relayer) initializeAndSync() error {
	return r.waitForBTCToSyncWithNative()
}

func (r *Relayer) setupCache(skipSubscription bool) error {
	if err := r.initializeBTCCache(); err != nil {
		return err
	}

	if !skipSubscription {
		r.btcClient.SubscribeNewBlocks()
	}
	return nil
}

func (r *Relayer) processAndTrimCache() error {
	if err := r.processHeaders(); err != nil {
		return err
	}

	err := r.resizeAndTrimCache()
	return err
}

func (r *Relayer) processHeaders() error {
	blocks := r.btcCache.GetAllBlocks()
	if _, err := r.ProcessHeaders(blocks); err != nil {
		// occurs when multiple competing spv relayers exist
		// or when our btc node is not fully synchronized
		r.logger.Errorf("Failed to submit headers: %v", err)
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
			return r.bootstrapRelayer(skipSubscription)
		},
		retryOpts...,
	); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		r.logger.Fatalf("Failed to bootstrap relayer: %v after %d attempts", err, bootstrapRetryAttempts)
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
			r.logger.Warnf(
				"Error bootstrapping relayer: %v. Attempts: %d, Max attempts: %d",
				err, n+1, bootstrapRetryAttempts,
			)
		}),
	}
}

// initializeBTCCache initializes the BTC cache with blocks from T-k to T
// where T is the height of the latest block in Native light client
// and k is the confirmation depth
func (r *Relayer) initializeBTCCache() error {
	cache, err := relayertypes.NewBTCCache(r.Config.BTCCacheSize)
	if err != nil {
		return err
	}
	r.btcCache = cache

	nativeBlockHeight, err := r.getNativeLatestBlockHeight()
	if err != nil {
		return err
	}

	baseHeight := nativeBlockHeight - r.btcConfirmationDepth + 1

	blocks, err := r.btcClient.GetBTCTailBlocksByHeight(baseHeight)
	if err != nil {
		return err
	}

	err = r.btcCache.Init(blocks)
	return err
}

// waitForBTCToSyncWithNative ensures BTC node is synchronized by checking
// that its chain height is at least equal to the Native light client height.
// This synchronization is required before proceeding with relayer operations.
func (r *Relayer) waitForBTCToSyncWithNative() error {
	btcLatestBlockHeight, err := r.getBTCLatestBlockHeight()
	if err != nil {
		return err
	}

	nativeLatestBlockHeight, err := r.getNativeLatestBlockHeight()
	if err != nil {
		return err
	}

	if btcLatestBlockHeight == 0 || btcLatestBlockHeight < nativeLatestBlockHeight {
		return r.waitForBTCCatchup(btcLatestBlockHeight, nativeLatestBlockHeight)
	}

	return nil
}

func (r *Relayer) getBTCLatestBlockHeight() (int64, error) {
	_, btcLatestBlockHeight, err := r.btcClient.GetBTCTipBlock()
	if err != nil {
		return 0, err
	}

	r.logger.Infof(
		"BTC latest block height: (%d)",
		btcLatestBlockHeight,
	)

	return btcLatestBlockHeight, nil
}

func (r *Relayer) getNativeLatestBlockHeight() (int64, error) {
	nativeBlock, err := r.nativeClient.GetHeaderChainTip()
	if err != nil {
		return 0, err
	}

	r.logger.Infof(
		"Light client header chain latest block height: (%d)",
		nativeBlock.Height,
	)

	return nativeBlock.Height, nil
}

func (r *Relayer) waitForBTCCatchup(btcHeight int64, nativeHeight int64) error {
	r.logger.Infof(
		"BTC chain (length %d) falls behind light client header chain (length %d), wait until BTC catches up",
		btcHeight, nativeHeight,
	)

	ticker := time.NewTicker(bootstrapSyncTicker)
	defer ticker.Stop()

	for {
		btcLatestBlockHeight, err := r.getBTCLatestBlockHeight()
		if err != nil {
			return err
		}

		nativeLatestBlockHeight, err := r.getNativeLatestBlockHeight()
		if err != nil {
			return err
		}

		if isBTCCaughtUp(btcLatestBlockHeight, nativeLatestBlockHeight) {
			r.logger.Infof(
				"BTC (height %d) has synchronized with light client header (height %d), proceeding with bootstrap",
				btcLatestBlockHeight, nativeLatestBlockHeight,
			)
			return nil
		}

		r.logger.Infof(
			"BTC (height %d) is not yet synchronized with light client header (height %d), continuing to wait",
			btcLatestBlockHeight, nativeLatestBlockHeight,
		)

		<-ticker.C
	}
}

func isBTCCaughtUp(btcHeight int64, nativeHeight int64) bool {
	return btcHeight > 0 && btcHeight >= nativeHeight
}
