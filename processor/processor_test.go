package processor

import (
	"context"
	"sync"
	"testing"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	"github.com/gonative-cc/relayer/ika"
	"github.com/stretchr/testify/assert"
)

func TestProcessor_ProcessTransaction(t *testing.T) {
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

	err := processor.ProcessTransaction(context.Background(), nativeTx, &sync.Mutex{})
	assert.Nil(t, err)

	retrievedTx, err := db.GetTx(nativeTx.TxID)
	assert.Nil(t, err)
	assert.Equal(t, retrievedTx.BtcTxID, uint64(nativeTx.TxID))
	assert.Equal(t, retrievedTx.Status, dal.StatusSigned)

}
