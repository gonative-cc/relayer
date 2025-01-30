package native2ika

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	"github.com/gonative-cc/relayer/ika"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newIkaProcessor creates a new Processor instance with a mocked IKA client and populated database.
func newIkaProcessor(t *testing.T, ikaClient ika.Client) *Processor {
	db := daltest.InitTestDB(t)
	return &Processor{
		ikaClient: ikaClient,
		db:        db,
	}
}

func TestRun(t *testing.T) {
	processor := newIkaProcessor(t, ika.NewMockClient())
	daltest.PopulateSignRequests(t, processor.db)
	daltest.PopulateBitcoinTxs(t, processor.db)

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

func TestRun_NoPendingRequests(t *testing.T) {
	mockIkaClient := new(ika.MockClient)
	processor := newIkaProcessor(t, mockIkaClient)

	err := processor.Run(context.Background())
	assert.Nil(t, err)
	assert.NoError(t, err)
}

func TestRun_IKAClientError(t *testing.T) {
	mockIkaClient := new(ika.MockClient)
	processor := newIkaProcessor(t, mockIkaClient)
	daltest.PopulateSignRequests(t, processor.db)

	mockIkaClient.On("ApproveAndSign", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, "", errors.New("ika client error"))

	err := processor.Run(context.Background())
	assert.ErrorContains(t, err, "failed calling ApproveAndSign")
}

func TestRun_Success(t *testing.T) {
	mockIkaClient := new(ika.MockClient)
	processor := newIkaProcessor(t, mockIkaClient)
	daltest.PopulateSignRequests(t, processor.db)

	signRequests, err := processor.db.GetPendingIkaSignRequests()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(signRequests))

	mockIkaClient.On("ApproveAndSign", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([][]byte{[]byte("signature")}, "txDigest", nil)

	err = processor.Run(context.Background())
	assert.NoError(t, err)

	signRequests, err = processor.db.GetPendingIkaSignRequests()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(signRequests)) // All requests should be processed
}

func TestRun_EmptyPayload(t *testing.T) {
	mockIkaClient := new(ika.MockClient)
	processor := newIkaProcessor(t, mockIkaClient)
	request := dal.IkaSignRequest{ID: 1, Payload: make([]byte, 0), DWalletID: "dwallet1",
		UserSig: "user_sig1", FinalSig: nil, Timestamp: time.Now().Unix()}

	err := processor.db.InsertIkaSignRequest(request)
	assert.NoError(t, err)

	mockIkaClient.On("ApproveAndSign", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([][]byte{[]byte("signature")}, "txDigest", nil)

	err = processor.Run(context.Background())
	assert.NoError(t, err)
}
