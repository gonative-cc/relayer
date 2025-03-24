package remote2ika

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
	ctx := context.Background()
	db := daltest.InitTestDB(ctx, t)
	return &Processor{
		ikaClient: ikaClient,
		db:        db,
	}
}

func newIkaMockWithApproveAndSingReq() ika.Client {
	suiCl := ika.NewMockClientWithApprove()
	// only signatures that don't have final sig
	suiCl.On("SignReq", mock.Anything, "dwallet1", "user_sig1", mock.Anything).
		Return("ds1", nil)
	suiCl.On("SignReq", mock.Anything, "dwallet3", "user_sig3", mock.Anything).
		Return("ds3", nil)

	return suiCl
}

func TestRun(t *testing.T) {
	ctx := context.Background()
	suiCl := newIkaMockWithApproveAndSingReq()
	processor := newIkaProcessor(t, suiCl)
	daltest.PopulateSignRequests(ctx, t, processor.db)
	daltest.PopulateBitcoinTxs(ctx, t, processor.db)

	// before signing
	retrievedSignRequests, err := processor.db.GetBitcoinTxsToBroadcast(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(retrievedSignRequests))
	assert.NotNil(t, retrievedSignRequests[0].FinalSig)

	err = processor.Run(ctx)
	assert.Nil(t, err)

	// after signing
	retrievedSignRequests, err = processor.db.GetBitcoinTxsToBroadcast(ctx)
	assert.Nil(t, err)

	// TODO
	t.Skip("Fetching Ika Tx is not ready - see TODO in ika/client.SignReq")
	assert.Equal(t, 3, len(retrievedSignRequests))
	assert.NotNil(t, retrievedSignRequests[0].FinalSig)
}

type testRunCase struct {
	ikaClient     ika.Client
	setupDB       func(t *testing.T, db dal.DB)
	assertions    func(t *testing.T, processor *Processor)
	name          string
	expectedError string
}

func TestRun_EdgeCases(t *testing.T) {
	// t.SkipNow()
	ctx := context.Background()
	testCases := []testRunCase{
		{
			name:      "NoPendingRequests",
			ikaClient: new(ika.MockClient),
			setupDB:   func(_ *testing.T, _ dal.DB) {}, // No need to populate the database
		},
		{
			name:      "Success",
			ikaClient: newIkaMockWithApproveAndSingReq(),
			setupDB: func(t *testing.T, db dal.DB) {
				daltest.PopulateSignRequests(ctx, t, db)
			},
			assertions: func(t *testing.T, processor *Processor) {
				signRequests, err := processor.db.GetPendingIkaSignRequests(ctx)
				assert.NoError(t, err)
				assert.NotNil(t, signRequests)
				// TODO: check if sign requests are processed
				// assert.Equal(t, 0, len(signRequests)) // All requests should be processed
			},
		},
		{
			name:      "EmptyPayload",
			ikaClient: newIkaMockWithApproveAndSingReq(),
			setupDB: func(t *testing.T, db dal.DB) {
				request := dal.IkaSignRequest{ID: 1, Payload: make([]byte, 0), DWalletID: "dwallet1",
					UserSig: "user_sig1", FinalSig: nil, Timestamp: time.Now().Unix()}
				err := db.InsertIkaSignRequest(ctx, request)
				assert.NoError(t, err)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			processor := newIkaProcessor(t, tc.ikaClient)
			tc.setupDB(t, processor.db)

			err := processor.Run(ctx)
			if tc.expectedError != "" {
				assert.ErrorContains(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			if tc.assertions != nil {
				tc.assertions(t, processor)
			}
		})
	}
}
