package bitcoinspv

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

var (
	bootstrapAttempts      = uint(60)
	bootstrapAttemptsAtt   = retry.Attempts(bootstrapAttempts)
	bootstrapRetryInterval = retry.Delay(30 * time.Second)
	bootstrapDelayType     = retry.DelayType(retry.FixedDelay)
	bootstrapErrReportType = retry.LastErrorOnly(true)
)

func (r *Relayer) bootstrapRelayer(skipBlockSubscription bool) error {
	// ensure BTC has caught up with Native header chain
	if err := r.waitForBTCToSyncWithNative(); err != nil {
		return err
	}

	// initialize cache with the latest blocks
	if err := r.initBTCCache(); err != nil {
		return err
	}
	r.logger.Debugf("BTC cache size: %d", r.btcCache.Size())

	// Subscribe new blocks after cache init to avoid duplicates
	if !skipBlockSubscription {
		r.btcClient.MustSubscribeBlocks()
	}

	// process headers from cache blocks
	blocks := r.btcCache.GetAllBlocks()
	if _, err := r.ProcessHeaders(blocks); err != nil {
		// this can happen when there are two contentious spv relayers or if our btc node is behind.
		r.logger.Errorf("failed to submit headers: %v", err)
		return err
	}

	// resize and trim cache to latest k blocks on BTC
	if err := r.resizeAndTrimCache(); err != nil {
		return err
	}

	r.logger.Infof("size of the BTC cache: %d", r.btcCache.Size())
	r.logger.Info("successfully finished bootstrapping")
	return nil
}

func (r *Relayer) resizeAndTrimCache() error {
	maxEntries := r.btcConfirmationDepth
	if err := r.btcCache.Resize(maxEntries); err != nil {
		r.logger.Errorf("failed to resize BTC cache: %v", err)
		panic(err)
	}
	r.btcCache.Trim()
	return nil
}

func (r *Relayer) createRelayerContext() (context.Context, func()) {
	quit := r.quitChan()
	ctx, cancel := context.WithCancel(context.Background())
	r.wg.Add(1)

	go func() {
		defer func() {
			cancel()
			r.wg.Done()
		}()

		select {
		case <-quit:

		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}

func (r *Relayer) multitryBootstrap(skipBlockSubscription bool) {
	ctx, cancel := r.createRelayerContext()
	defer cancel()

	retryOpts := []retry.Option{
		retry.Context(ctx),
		bootstrapAttemptsAtt,
		bootstrapRetryInterval,
		bootstrapDelayType,
		bootstrapErrReportType,
		retry.OnRetry(func(n uint, err error) {
			r.logger.Warnf(
				"failed to bootstap relayer: %v. Attempt: %d, Max attempts: %d",
				err, n+1, bootstrapAttempts,
			)
		}),
	}

	if err := retry.Do(
		func() error {
			return r.bootstrapRelayer(skipBlockSubscription)
		},
		retryOpts...,
	); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		r.logger.Fatalf("failed to bootstrap relayer: %v after %d attempts", err, bootstrapAttempts)
	}
}

// initBTCCache initializes the BTC cache with blocks from T-k to T
// where T is the height of the latest block in Native light client
// and k is the confirmation depth
func (r *Relayer) initBTCCache() error {
	// Initialize cache with configured size
	cache, err := types.NewBTCCache(r.Cfg.BTCCacheSize) // TODO: give an option to be unsized
	if err != nil {
		panic(err)
	}
	r.btcCache = cache

	// Get latest block height from Native chain
	nativeBlockHeight, err := r.getNativeLatestBlockHeight()
	if err != nil {
		return err
	}

	// Calculate base height to fetch blocks from
	baseHeight := nativeBlockHeight - r.btcConfirmationDepth + 1

	// Fetch blocks from BTC node
	blocks, err := r.btcClient.GetTailBlocksByHeight(baseHeight)
	if err != nil {
		panic(err)
	}

	// Initialize cache with fetched blocks
	if err = r.btcCache.Init(blocks); err != nil {
		panic(err)
	}
	return nil
}

// waitForBTCToSyncWithNative waits for BTC to synchronize until BTC is no shorter than Native's BTC light client.
// It returns BTC last block hash, BTC last block height, and Native's base height.
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
	btcLatestBlockHash, btcLatestBlockHeight, err := r.btcClient.GetBestBlock()
	if err != nil {
		return 0, err
	}

	r.logger.Infof(
		"BTC latest block hash and height: (%v, %d)", btcLatestBlockHash, btcLatestBlockHeight,
	)

	return btcLatestBlockHeight, nil
}

func (r *Relayer) getNativeLatestBlockHeight() (int64, error) {
	nativeBlock, err := r.nativeClient.GetBTCHeaderChainTip()
	if err != nil {
		return 0, err
	}

	r.logger.Infof(
		"Light client header chain latest block hash and height: (%v, %d)",
		nativeBlock.Hash, nativeBlock.Height,
	)

	return nativeBlock.Height, nil
}

func (r *Relayer) waitForBTCCatchup(btcHeight, nativeHeight int64) error {
	r.logger.Infof(
		"BTC chain (length %d) falls behind light client header chain (length %d), wait until BTC catches up",
		btcHeight, nativeHeight,
	)

	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		btcLatestBlockHeight, err := r.getBTCLatestBlockHeight()
		if err != nil {
			return err
		}

		nativeLatestBlockHeight, err := r.getNativeLatestBlockHeight()
		if err != nil {
			return err
		}

		if btcLatestBlockHeight > 0 && btcLatestBlockHeight >= nativeLatestBlockHeight {
			r.logger.Infof(
				"BTC chain (length %d) now catches up with light client header chain (length %d), continue bootstrapping",
				btcLatestBlockHeight, nativeLatestBlockHeight,
			)
			break
		}

		r.logger.Infof(
			"BTC chain (length %d) still falls behind light client header chain (length %d), keep waiting",
			btcLatestBlockHeight, nativeLatestBlockHeight,
		)
	}

	return nil
}
