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
	ctx := context.Background()
	processor := newIkaProcessor(t, ika.NewMockClient())
	daltest.PopulateSignRequests(ctx, t, processor.db)
	daltest.PopulateBitcoinTxs(ctx, t, processor.db)

	// before signing
	retrievedSignRequests, err := processor.db.GetBitcoinTxsToBroadcast(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(retrievedSignRequests))
	assert.NotNil(t, retrievedSignRequests[0].FinalSig)

	err = processor.Run(context.Background())
	assert.Nil(t, err)

	// after signing
	retrievedSignRequests, err = processor.db.GetBitcoinTxsToBroadcast(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(retrievedSignRequests))
	assert.NotNil(t, retrievedSignRequests[0].FinalSig)
}

type testRunCase struct {
	name          string
	ikaClient     ika.Client
	setupDB       func(t *testing.T, db dal.DB)
	expectedError string
	assertions    func(t *testing.T, processor *Processor)
}

func TestRun_EdgeCases(t *testing.T) {
	ctx := context.Background()
	testCases := []testRunCase{
		{
			name:      "NoPendingRequests",
			ikaClient: new(ika.MockClient),
			setupDB:   func(_ *testing.T, _ dal.DB) {}, // No need to populate the database
		},
		{
			name: "IKAClientError",
			ikaClient: func() ika.Client {
				mockIkaClient := new(ika.MockClient)
				mockIkaClient.On("ApproveAndSign", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, "", errors.New("ika client error"))
				return mockIkaClient
			}(),
			setupDB: func(t *testing.T, db dal.DB) {
				daltest.PopulateSignRequests(ctx, t, db)
			},
			expectedError: "failed calling ApproveAndSign",
		},
		{
			name: "Success",
			ikaClient: func() ika.Client {
				mockIkaClient := new(ika.MockClient)
				mockIkaClient.On("ApproveAndSign", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([][]byte{[]byte("signature")}, "txDigest", nil)
				return mockIkaClient
			}(),
			setupDB: func(t *testing.T, db dal.DB) {
				daltest.PopulateSignRequests(ctx, t, db)
			},
			assertions: func(t *testing.T, processor *Processor) {
				signRequests, err := processor.db.GetPendingIkaSignRequests(ctx)
				assert.NoError(t, err)
				assert.Equal(t, 0, len(signRequests)) // All requests should be processed
			},
		},
		{
			name: "EmptyPayload",
			ikaClient: func() ika.Client {
				mockIkaClient := new(ika.MockClient)
				mockIkaClient.On("ApproveAndSign", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([][]byte{[]byte("signature")}, "txDigest", nil)
				return mockIkaClient
			}(),
			setupDB: func(t *testing.T, db dal.DB) {
				request := dal.IkaSignRequest{ID: 1, Payload: make([]byte, 0), DwalletID: "dwallet1",
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

			err := processor.Run(context.Background())
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
