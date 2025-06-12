package bitcoinspv

import (
	"context"
	"errors"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBootstrapRelayer(t *testing.T) {
	ctx := context.Background()
	const latestHeight = 90
	const latestFinalized = latestHeight - confirmationDepth + 1

	t.Run("successful bootstrap", func(t *testing.T) {
		r, btcClient, lcClient := setupTest(t)

		btcClient.On("GetBTCTipBlock").Return(&chainhash.Hash{}, int64(latestHeight), nil)
		lcClient.On("GetLatestBlockInfo", ctx).Return(&clients.BlockInfo{
			Height: latestHeight,
		}, nil)
		blocks := make([]*types.IndexedBlock, confirmationDepth)
		for i := 0; i < confirmationDepth; i++ {
			blocks[i] = &types.IndexedBlock{
				RawMsgBlock: wire.NewMsgBlock(&wire.BlockHeader{}),
				BlockHeight: int64(i + latestFinalized),
			}
		}
		btcClient.On("GetBTCTailBlocksByHeight", int64(latestFinalized), false).Return(blocks, nil)
		lcClient.On("ContainsBlock", mock.Anything, mock.Anything).Return(true, nil)
		btcClient.On("SubscribeNewBlocks").Return()

		err := r.bootstrapRelayer(ctx, false)
		assert.NoError(t, err)
	})

	t.Run("failed BTC sync", func(t *testing.T) {
		r, btcClient, _ := setupTest(t)

		expectedErr := errors.New("btc sync failed")
		btcClient.On("GetBTCTipBlock").Return(nil, int64(0), expectedErr)

		err := r.bootstrapRelayer(ctx, false)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("failed light client sync", func(t *testing.T) {
		r, btcClient, lcClient := setupTest(t)
		btcClient.On("GetBTCTipBlock").Return(&chainhash.Hash{}, int64(latestFinalized), nil)

		expectedErr := errors.New("light client sync failed")
		lcClient.On("GetLatestBlockInfo", mock.Anything).Return(nil, expectedErr)

		err := r.bootstrapRelayer(ctx, false)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestWaitForBTCCatchup(t *testing.T) {
	r, btcClient, lcClient := setupTest(t)
	ctx := context.Background()

	// First call: BTC behind
	btcClient.On("GetBTCTipBlock").Return(&chainhash.Hash{}, int64(90), nil).Twice()
	// Second call: BTC caught up
	btcClient.On("GetBTCTipBlock").Return(&chainhash.Hash{}, int64(95), nil).Once()

	// LC height stays constant
	lcClient.On("GetLatestBlockInfo", ctx).Return(&clients.BlockInfo{
		Height: 95,
	}, nil)

	// LC height stays constant
	lcClient.On("GetLatestBlockInfo", ctx).Return(&clients.BlockInfo{
		Height: 95,
	}, nil)

	err := r.waitForBitcoinCatchup(ctx)
	assert.NoError(t, err)
}
