package bitcoinspv

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

var (
	bootstrapAttempts      = uint(60)
	bootstrapAttemptsAtt   = retry.Attempts(bootstrapAttempts)
	bootstrapRetryInterval = retry.Delay(30 * time.Second)
	bootstrapDelayType     = retry.DelayType(retry.FixedDelay)
	bootstrapErrReportType = retry.LastErrorOnly(true)
)

func (r *Relayer) bootstrap(skipBlockSubscription bool) error {
	var (
		ibs []*types.IndexedBlock
		err error
	)

	// ensure BTC has caught up with Native header chain
	if err := r.waitUntilBTCSync(); err != nil {
		return err
	}

	// initialize cache with the latest blocks
	if err := r.initBTCCache(); err != nil {
		return err
	}
	r.logger.Debugf("BTC cache size: %d", r.btcCache.Size())

	// Subscribe new blocks right after initializing BTC cache,
	// in order to ensure subscribed blocks and cached blocks do not have overlap.
	// Otherwise, if we subscribe too early, then they will have overlap,
	// leading to duplicated header submissions.
	if !skipBlockSubscription {
		r.btcClient.MustSubscribeBlocks()
	}

	ibs = r.btcCache.GetAllBlocks()

	// r.logger.Infof(
	// 	"BTC height: %d. BTCLightclient height: %d. Start syncing from height %d.",
	// 	btcLatestBlockHeight, consistencyInfo.nativeLatestBlockHeight, consistencyInfo.startSyncHeight,
	// )

	// extracts and submits headers for each block in ibs
	// Note: As we are retrieving blocks from btc cache from block just after confirmed block which
	// we already checked for consistency, we can be sure that
	// even if rest of the block headers is different than in native lightclient
	// due to reorg, our fork will be better than the one in native lightclient.
	_, err = r.ProcessHeaders(ibs)
	if err != nil {
		// this can happen when there are two contentious spv relayers or if our btc node is behind.
		r.logger.Errorf("Failed to submit headers: %v", err)
		// returning error as it is up to the caller to decide what do next
		return err
	}

	// trim cache to the latest k blocks on BTC (which are same as in Native)
	maxEntries := r.btcConfirmationDepth
	if err = r.btcCache.Resize(maxEntries); err != nil {
		r.logger.Errorf("Failed to resize BTC cache: %v", err)
		panic(err)
	}
	r.btcCache.Trim()

	r.logger.Infof("Size of the BTC cache: %d", r.btcCache.Size())

	r.logger.Info("Successfully finished bootstrapping")
	return nil
}

func (r *Relayer) relayerQuitCtx() (context.Context, func()) {
	quit := r.quitChan()
	ctx, cancel := context.WithCancel(context.Background())
	r.wg.Add(1)
	go func() {
		defer cancel()
		defer r.wg.Done()

		select {
		case <-quit:

		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}

func (r *Relayer) bootstrapWithRetries(skipBlockSubscription bool) {
	// if we are exiting, we need to cancel this process
	ctx, cancel := r.relayerQuitCtx()
	defer cancel()
	if err := retry.Do(func() error {
		return r.bootstrap(skipBlockSubscription)
	},
		retry.Context(ctx),
		bootstrapAttemptsAtt,
		bootstrapRetryInterval,
		bootstrapDelayType,
		bootstrapErrReportType, retry.OnRetry(func(n uint, err error) {
			r.logger.Warnf(
				"Failed to bootstap relayer: %v. Attempt: %d, Max attempts: %d",
				err, n+1, bootstrapAttempts,
			)
		})); err != nil {

		if errors.Is(err, context.Canceled) {
			// context was canceled we do not need to anything more, app is quiting
			return
		}

		// we failed to bootstrap multiple time, we should panic as something unexpected is happening.
		r.logger.Fatalf("Failed to bootstrap relayer: %v after %d attempts", err, bootstrapAttempts)
	}
}

// initBTCCache fetches the blocks since T-k-w in the BTC canonical chain
// where T is the height of the latest block in Native light client
func (r *Relayer) initBTCCache() error {
	var (
		err                     error
		nativeLatestBlockHeight int64
		baseHeight              int64
		ibs                     []*types.IndexedBlock
	)

	r.btcCache, err = types.NewBTCCache(r.Cfg.BTCCacheSize) // TODO: give an option to be unsized
	if err != nil {
		panic(err)
	}

	// get T, i.e., total block count in Native light client
	chainBlock, err := r.nativeClient.GetBTCHeaderChainTip()
	if err != nil {
		return err
	}
	nativeLatestBlockHeight = chainBlock.Height

	// Fetch block since `baseHeight = T - k` from BTC, where
	// - T is total block count in Native light client
	// - k is btcConfirmationDepth of Native
	baseHeight = nativeLatestBlockHeight - r.btcConfirmationDepth + 1

	ibs, err = r.btcClient.FindTailBlocksByHeight(baseHeight)
	if err != nil {
		panic(err)
	}

	if err = r.btcCache.Init(ibs); err != nil {
		panic(err)
	}
	return nil
}

// waitUntilBTCSync waits for BTC to synchronize until BTC is no shorter than Native's BTC light client.
// It returns BTC last block hash, BTC last block height, and Native's base height.
func (r *Relayer) waitUntilBTCSync() error {
	var (
		btcLatestBlockHash      *chainhash.Hash
		btcLatestBlockHeight    int64
		nativeLatestBlockHash   *chainhash.Hash
		nativeLatestBlockHeight int64
		err                     error
	)

	// Retrieve hash/height of the latest block in BTC
	btcLatestBlockHash, btcLatestBlockHeight, err = r.btcClient.GetBestBlock()
	if err != nil {
		return err
	}
	r.logger.Infof(
		"BTC latest block hash and height: (%v, %d)", btcLatestBlockHash, btcLatestBlockHeight,
	)

	// TODO: if BTC falls behind BTCLightclient's base header,
	// then the spv relayer is incorrectly configured and should panic

	// Retrieve hash/height of the latest block in Native header chain
	chainBlock, err := r.nativeClient.GetBTCHeaderChainTip()
	if err != nil {
		return err
	}

	// hash, err := types.NewBTCHeaderHashBytesFromHex(tipRes.Header.HashHex)
	// if err != nil {
	// 	return err
	// }

	nativeLatestBlockHash = chainBlock.Hash
	nativeLatestBlockHeight = chainBlock.Height
	r.logger.Infof(
		"Light client header chain latest block hash and height: (%v, %d)",
		nativeLatestBlockHash, nativeLatestBlockHeight,
	)

	// If BTC chain is shorter than Native header chain, pause until BTC catches up
	if btcLatestBlockHeight == 0 || btcLatestBlockHeight < nativeLatestBlockHeight {
		r.logger.Infof(
			"BTC chain (length %d) falls behind light client header chain (length %d), wait until BTC catches up",
			btcLatestBlockHeight, nativeLatestBlockHeight,
		)

		// periodically check if BTC catches up with Native.
		// When BTC catches up, break and continue the bootstrapping process
		ticker := time.NewTicker(5 * time.Second) // TODO: parameterise the polling interval
		for range ticker.C {
			_, btcLatestBlockHeight, err = r.btcClient.GetBestBlock()
			if err != nil {
				return err
			}
			chainBlock, err := r.nativeClient.GetBTCHeaderChainTip()
			if err != nil {
				return err
			}
			nativeLatestBlockHeight = chainBlock.Height
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
	}

	return nil
}
