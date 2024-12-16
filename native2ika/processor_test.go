package native2ika

import (
	"context"
	"sync"
	"testing"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	"github.com/gonative-cc/relayer/ika"
	"github.com/stretchr/testify/assert"
)

func TestProcessor_ProcessTxs(t *testing.T) {
	db := daltest.InitTestDB(t)

	mockIkaClient := ika.NewMockClient()

	processor := &Processor{
		ikaClient: mockIkaClient,
		db:        db,
	}

	nativeTx := &NativeTxData{
		TxID:           123,
		DWalletCapID:   "capID",
		SignMessagesID: "msgID",
		Messages:       [][]byte{{1, 2, 3}},
	}

	err := processor.ProcessTxs(context.Background(), nativeTx, &sync.Mutex{})
	assert.Nil(t, err)

	retrievedTx, err := db.GetTx(nativeTx.TxID)
	assert.Nil(t, err)
	assert.Equal(t, retrievedTx.BtcTxID, uint64(nativeTx.TxID))
	assert.Equal(t, retrievedTx.Status, dal.StatusSigned)

}

func TestProcessor_ProcessTxs_Multiple(t *testing.T) {
	db := daltest.InitTestDB(t)

	mockIkaClient := ika.NewMockClient()

	processor := &Processor{
		ikaClient: mockIkaClient,
		db:        db,
	}
	nativeTx := &NativeTxData{
		TxID:           123,
		DWalletCapID:   "capID",
		SignMessagesID: "msgID",
		Messages:       [][]byte{{1, 1, 1}, {2, 2, 2}, {3, 3, 3}},
	}

	err := processor.ProcessTxs(context.Background(), nativeTx, &sync.Mutex{})
	assert.Nil(t, err)

	retrievedTx, err := db.GetTx(nativeTx.TxID)
	assert.Nil(t, err)
	assert.Equal(t, retrievedTx.BtcTxID, uint64(nativeTx.TxID))
	assert.Equal(t, retrievedTx.Status, dal.StatusSigned)
}
