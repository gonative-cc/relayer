package ika2btc

import (
	"sync"
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

func TestProcessor_ProcessSignedTxs(t *testing.T) {
	db := daltest.InitTestDB(t)
	daltest.PopulateSignRequests(t, db)
	daltest.PopulateBitcoinTxs(t, db)
	processor, err := NewProcessor(btcClientConfig, 6, db)
	assert.NilError(t, err)
	processor.BtcClient = &bitcoin.MockClient{}

	var mu sync.Mutex
	err = processor.ProcessSignedTxs(&mu)
	assert.NilError(t, err)

	updatedTx, err := db.GetBitcoinTx(2, daltest.GetHashBytes(t, "0"))
	assert.NilError(t, err)
	assert.Equal(t, updatedTx.Status, dal.Broadcasted)
}

func TestProcessor_CheckConfirmations(t *testing.T) {
	db := daltest.InitTestDB(t)
	daltest.PopulateSignRequests(t, db)
	daltest.PopulateBitcoinTxs(t, db)
	processor, err := NewProcessor(btcClientConfig, 6, db)
	assert.NilError(t, err)
	processor.BtcClient = &bitcoin.MockClient{}

	var mu sync.Mutex
	err = processor.CheckConfirmations(&mu)
	assert.NilError(t, err)

	updatedTx, err := db.GetBitcoinTx(4, daltest.GetHashBytes(t, "3"))
	assert.NilError(t, err)
	assert.Equal(t, updatedTx.Status, dal.Confirmed)
}

func TestNewProcessor_DatabaseError(t *testing.T) {
	processor, err := NewProcessor(btcClientConfig, 6, nil)
	assert.ErrorContains(t, err, "database cannot be nil")
	assert.Assert(t, processor == nil)
}

func TestNewProcessor_MissingBtcConfig(t *testing.T) {
	db := daltest.InitTestDB(t)
	btcClientConfig.Host = ""
	processor, err := NewProcessor(btcClientConfig, 6, db)
	assert.ErrorContains(t, err, "missing bitcoin node configuration")
	assert.Assert(t, processor == nil)
}
