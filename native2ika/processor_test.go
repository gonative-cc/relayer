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

// newIkaProcessor creates a new Processor instance with a mocked IKA client and populated database.
func newIkaProcessor(t *testing.T) *Processor {
	db := daltest.InitTestDB(t)
	daltest.PopulateNativeDB(t, db)

	mockIkaClient := ika.NewMockClient()

	processor := &Processor{
		ikaClient: mockIkaClient,
		db:        db,
	}

	return processor
}

func TestProcessor_ProcessTxs(t *testing.T) {
	processor := newIkaProcessor(t)

	err := processor.ProcessTxs(context.Background(), &sync.Mutex{})
	assert.Nil(t, err)

	retrievedTxs, err := processor.db.GetSignedTxs()
	assert.Nil(t, err)
	assert.Equal(t, len(retrievedTxs), 2)
	assert.Equal(t, retrievedTxs[0].Status, dal.StatusSigned)
}
