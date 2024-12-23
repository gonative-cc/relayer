package ika2btc

import (
	"testing"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	"github.com/gonative-cc/relayer/errors"
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
	processor, db := initProcessor(t)
	err := processor.Run()
	assert.NilError(t, err)

	updatedTx, err := db.GetBitcoinTx(2, daltest.GetHashBytes(t, "0"))
	assert.NilError(t, err)
	assert.Equal(t, updatedTx.Status, dal.Broadcasted)
}

func TestCheckConfirmations(t *testing.T) {
	processor, db := initProcessor(t)
	processor.BtcClient = &bitcoin.MockClient{}
	err := processor.CheckConfirmations()
	assert.NilError(t, err)

	updatedTx, err := db.GetBitcoinTx(4, daltest.GetHashBytes(t, "3"))
	assert.NilError(t, err)
	assert.Equal(t, updatedTx.Status, dal.Confirmed)
}

func TestNewProcessor(t *testing.T) {
	// missing db
	processor, err := NewProcessor(btcClientConfig, 6, nil)
	assert.ErrorIs(t, err, errors.ErrNoDB)
	assert.Assert(t, processor == nil)

	// missing BTC config
	db := daltest.InitTestDB(t)
	btcClientConfig.Host = ""
	processor, err = NewProcessor(btcClientConfig, 6, db)
	assert.ErrorIs(t, err, errors.ErrNoBtcConfig)
	assert.Assert(t, processor == nil)
}

// initProcessor initializes processor with a mock Bitcoin client and a populated database.
func initProcessor(t *testing.T) (*Processor, *dal.DB) {
	t.Helper()

	db := daltest.InitTestDB(t)
	daltest.PopulateSignRequests(t, db)
	daltest.PopulateBitcoinTxs(t, db)
	processor, err := NewProcessor(btcClientConfig, 6, db)
	assert.NilError(t, err)
	processor.BtcClient = &bitcoin.MockClient{}

	return processor, db
}
