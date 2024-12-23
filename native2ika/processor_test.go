package native2ika

import (
	"context"
	"testing"

	"github.com/gonative-cc/relayer/dal/daltest"
	"github.com/gonative-cc/relayer/ika"
	"github.com/stretchr/testify/assert"
)

// newIkaProcessor creates a new Processor instance with a mocked IKA client and populated database.
func newIkaProcessor(t *testing.T) *Processor {
	db := daltest.InitTestDB(t)
	daltest.PopulateSignRequests(t, db)
	daltest.PopulateBitcoinTxs(t, db)

	mockIkaClient := ika.NewMockClient()

	return &Processor{
		ikaClient: mockIkaClient,
		db:        db,
	}
}

func TestRun(t *testing.T) {
	processor := newIkaProcessor(t)

	// before signing
	retrievedSignRequests, err := processor.db.GetBitcoinTxsToBroadcast()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(retrievedSignRequests))
	assert.NotNil(t, retrievedSignRequests[0].FinalSig)

	err = processor.Run(context.Background())
	assert.Nil(t, err)

	// after signing
	retrievedSignRequests, err = processor.db.GetBitcoinTxsToBroadcast()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(retrievedSignRequests))
	assert.NotNil(t, retrievedSignRequests[0].FinalSig)
}
