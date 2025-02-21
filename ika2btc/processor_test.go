package ika2btc

import (
	"context"
	"testing"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	"gotest.tools/v3/assert"
)

var btcClientConfig = rpcclient.ConnConfig{
	Host:         "test_rpc",
	User:         "test_user",
	Pass:         "test_pass",
	HTTPPostMode: true,
	DisableTLS:   false,
}

func TestProcessSignedTxs(t *testing.T) {
	ctx := context.Background()
	processor, db := initProcessor(t)
	err := processor.Run(ctx)
	assert.NilError(t, err)

	updatedTx, err := db.GetBitcoinTx(ctx, 2, daltest.DecodeBTCHash(t, "0"))
	assert.NilError(t, err)
	assert.Equal(t, updatedTx.Status, int64(dal.Broadcasted))
}

func TestCheckConfirmations(t *testing.T) {
	ctx := context.Background()
	processor, db := initProcessor(t)
	processor.BtcClient = &bitcoin.MockClient{}
	err := processor.CheckConfirmations(ctx)
	assert.NilError(t, err)

	updatedTx, err := db.GetBitcoinTx(ctx, 4, daltest.DecodeBTCHash(t, "3"))
	assert.NilError(t, err)
	assert.Equal(t, updatedTx.Status, int64(dal.Confirmed))
}

func TestNewProcessor(t *testing.T) {
	// missing BTC config
	ctx := context.Background()
	db := daltest.InitTestDB(ctx, t)
	btcClientConfig.Host = ""
	processor, err := NewProcessor(btcClientConfig, 6, db)
	assert.ErrorIs(t, err, bitcoin.ErrNoBtcConfig)
	assert.Assert(t, processor == nil)
}

// initProcessor initializes processor with a mock Bitcoin client and a populated database.
func initProcessor(t *testing.T) (*Processor, dal.DB) {
	t.Helper()
	ctx := context.Background()
	db := daltest.InitTestDB(ctx, t)
	daltest.PopulateSignRequests(ctx, t, db)
	daltest.PopulateBitcoinTxs(ctx, t, db)
	processor, err := NewProcessor(btcClientConfig, 6, db)
	assert.NilError(t, err)
	processor.BtcClient = &bitcoin.MockClient{}

	return processor, db
}
