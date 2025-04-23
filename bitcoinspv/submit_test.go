package bitcoinspv

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/clients/mocks"
	sui_errors "github.com/gonative-cc/relayer/bitcoinspv/clients/sui"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var testSubmitConfig = &config.RelayerConfig{
	RetrySleepDuration:    1 * time.Millisecond,
	MaxRetrySleepDuration: 50 * time.Millisecond,
	HeadersChunkSize:      2,
	Format:                "auto",
	Level:                 "debug",
	NetParams:             "test",
	BTCCacheSize:          100,
	BTCConfirmationDepth:  6,
	ProcessBlockTimeout:   5 * time.Second,
}

func TestFindFirstUnknownHeaderIndex(t *testing.T) {
	ctx := context.Background()
	testBlocks := types.CreateTestIndexedBlocks(t, 5, 100) // heights 100, 101, 102, 103, 104

	tests := []struct {
		name          string
		mockSetup     func(mockLC *mocks.MockBitcoinSPV)
		expectedIndex int
		expectedCalls int
		expectedErr   error
	}{
		{
			name: "all headers exist",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				for _, b := range testBlocks {
					mockLC.On("ContainsBlock", ctx, b.BlockHash()).Return(true, nil).Once()
				}
			},
			expectedIndex: -1,
			expectedCalls: 5,
		},
		{
			name: "no headers exist",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				mockLC.On("ContainsBlock", ctx, testBlocks[0].BlockHash()).Return(false, nil).Once()
			},
			expectedIndex: 0,
			expectedCalls: 1,
		},
		{
			name: "some headers exist",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				mockLC.On("ContainsBlock", ctx, testBlocks[0].BlockHash()).Return(true, nil).Once()
				mockLC.On("ContainsBlock", ctx, testBlocks[1].BlockHash()).Return(true, nil).Once()
				mockLC.On("ContainsBlock", ctx, testBlocks[2].BlockHash()).Return(false, nil).Once()
			},
			expectedIndex: 2,
			expectedCalls: 3,
		},
		{
			name: "ContainsBlock retryable error then success (finds new)",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				retryableErr := errors.New("network timeout")
				mockLC.On("ContainsBlock", ctx, testBlocks[0].BlockHash()).Return(true, nil).Once()
				mockLC.On("ContainsBlock", ctx, testBlocks[1].BlockHash()).Return(false, retryableErr).Once() // retry after error
				mockLC.On("ContainsBlock", ctx, testBlocks[1].BlockHash()).Return(true, nil).Once()           // success on retry
				mockLC.On("ContainsBlock", ctx, testBlocks[2].BlockHash()).Return(false, nil).Once()
			},
			expectedIndex: 2,
			expectedCalls: 4,
		},
		{
			name: "ContainsBlock abort error",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				abortErr := fmt.Errorf("%w: ... MoveAbort(...)", sui_errors.ErrSuiTransactionFailed)
				mockLC.On("ContainsBlock", ctx, testBlocks[0].BlockHash()).Return(false, abortErr).Once()
			},
			expectedIndex: -1,
			expectedCalls: 1,
			expectedErr:   fmt.Errorf("%w: ... MoveAbort(...)", sui_errors.ErrSuiTransactionFailed),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLC := mocks.NewMockBitcoinSPV(t)
			tt.mockSetup(mockLC)
			r := &Relayer{
				lcClient: mockLC,
				logger:   zerolog.Nop(),
				Config:   testSubmitConfig,
			}
			idx, err := r.FindFirstUnknownHeaderIndex(ctx, testBlocks)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				assert.Equal(t, tt.expectedIndex, idx, "Index should be -1 on error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedIndex, idx)
			}
			mockLC.AssertExpectations(t)
			mockLC.AssertNumberOfCalls(t, "ContainsBlock", tt.expectedCalls)
		})
	}
}

func TestCreateChunks(t *testing.T) {
	ctx := context.Background()
	testBlocks := types.CreateTestIndexedBlocks(t, 5, 100) // heights 100, 101, 102, 103, 104

	tests := []struct {
		name             string
		mockSetup        func(mockLC *mocks.MockBitcoinSPV)
		expectedChunkLen []int // len of headers in each chunk
	}{
		{
			name: "all headers new",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				mockLC.On("ContainsBlock", ctx, testBlocks[0].BlockHash()).Return(false, nil).Once()
			},
			expectedChunkLen: []int{2, 2, 1}, // {100, 101}{102, 103}{104}
		},
		{
			name: "some headers new (index 2)",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				mockLC.On("ContainsBlock", ctx, testBlocks[0].BlockHash()).Return(true, nil).Once()
				mockLC.On("ContainsBlock", ctx, testBlocks[1].BlockHash()).Return(true, nil).Once()
				mockLC.On("ContainsBlock", ctx, testBlocks[2].BlockHash()).Return(false, nil).Once()
			},
			expectedChunkLen: []int{2, 1}, // {102, 103}{104}
		},
		{
			name: "all headers exist",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				for _, b := range testBlocks {
					mockLC.On("ContainsBlock", ctx, b.BlockHash()).Return(true, nil).Once()
				}
			},
			expectedChunkLen: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLC := mocks.NewMockBitcoinSPV(t)
			tt.mockSetup(mockLC)
			r := &Relayer{
				lcClient: mockLC,
				logger:   zerolog.Nop(),
				Config:   testSubmitConfig,
			}
			chunks, err := r.createChunks(ctx, testBlocks)
			assert.NoError(t, err)
			if tt.expectedChunkLen == nil {
				assert.Empty(t, chunks)
			} else {
				assert.NotNil(t, chunks)
				assert.Len(t, chunks, len(tt.expectedChunkLen))
				for i, expectedLen := range tt.expectedChunkLen {
					assert.Len(t, chunks[i].Headers, expectedLen, "Header count in chunk %d mismatch", i)
				}
			}

			mockLC.AssertExpectations(t)
		})
	}
}

func TestSubmitHeaderMessages(t *testing.T) {
	ctx := context.Background()
	testChunk := Chunk{
		Headers: []wire.BlockHeader{{Version: 1}, {Version: 2}},
		From:    100,
		To:      101,
	}

	retryableErr := errors.New("network timeout")
	nonRetryableMoveAbortErr := fmt.Errorf("%w: function 'test_abort' status: failure, error: MoveAbort(MoveLocation { module: ..., name: ... }, 1234567890) in command 0",
		sui_errors.ErrSuiTransactionFailed)
	nonRetryableOutOfGasErr := fmt.Errorf("%w: function 'test_gas' status: failure, error: OutOfGas",
		sui_errors.ErrSuiTransactionFailed)

	tests := []struct {
		name          string
		mockSetup     func(mockLC *mocks.MockBitcoinSPV)
		expectedErr   bool
		expectedCalls int
	}{
		{
			name: "success on first try",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				mockLC.On("InsertHeaders", ctx, testChunk.Headers).Return(nil).Once()
			},
			expectedErr:   false,
			expectedCalls: 1,
		},
		{
			name: "retryable error then success",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				mockLC.On("InsertHeaders", ctx, testChunk.Headers).Return(retryableErr).Once()
				mockLC.On("InsertHeaders", ctx, testChunk.Headers).Return(nil).Once()
			},
			expectedErr:   false,
			expectedCalls: 2,
		},
		{
			name: "non-retryable MoveAbort error",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				mockLC.On("InsertHeaders", ctx, testChunk.Headers).Return(nonRetryableMoveAbortErr).Once()
			},
			expectedErr:   true,
			expectedCalls: 1,
		},
		{
			name: "non-retryable OutOfGas error",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				mockLC.On("InsertHeaders", ctx, testChunk.Headers).Return(nonRetryableOutOfGasErr).Once()
			},
			expectedErr:   true,
			expectedCalls: 1,
		},
		{
			name: "retryable error timeout",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				// This simulates RetryDo hitting timeout
				mockLC.On("InsertHeaders", ctx, testChunk.Headers).Return(retryableErr)
			},
			expectedErr:   true,
			expectedCalls: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLC := mocks.NewMockBitcoinSPV(t)
			tt.mockSetup(mockLC)
			r := &Relayer{
				lcClient: mockLC,
				logger:   zerolog.Nop(),
				Config:   testSubmitConfig,
			}
			err := r.submitHeaderMessages(ctx, testChunk)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockLC.AssertExpectations(t)
			if tt.expectedCalls > 0 {
				mockLC.AssertNumberOfCalls(t, "InsertHeaders", tt.expectedCalls)
			}
		})
	}
}
func TestProcessHeaders(t *testing.T) {
	ctx := context.Background()
	testBlocks := types.CreateTestIndexedBlocks(t, 5, 100) // heights 100, 101, 102, 103, 104

	expectedHeadersChunk1 := []wire.BlockHeader{*testBlocks[0].BlockHeader, *testBlocks[1].BlockHeader} // 100, 101
	expectedHeadersChunk2 := []wire.BlockHeader{*testBlocks[2].BlockHeader, *testBlocks[3].BlockHeader} // 102, 103
	expectedHeadersChunk3 := []wire.BlockHeader{*testBlocks[4].BlockHeader}                             // 104

	tests := []struct {
		name          string
		mockSetup     func(mockLC *mocks.MockBitcoinSPV)
		expectedCount int
		expectedErr   bool
	}{
		{
			name: "success - all headers new",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				// findFirstNewHeader finds index 0
				mockLC.On("ContainsBlock", ctx, testBlocks[0].BlockHash()).Return(false, nil).Once()
				// then submitHeaderMessages calls InsertHeaders
				mockLC.On("InsertHeaders", ctx, expectedHeadersChunk1).Return(nil).Once()
				mockLC.On("InsertHeaders", ctx, expectedHeadersChunk2).Return(nil).Once()
				mockLC.On("InsertHeaders", ctx, expectedHeadersChunk3).Return(nil).Once()
			},
			expectedCount: 5,
			expectedErr:   false,
		},
		{
			name: "success - some headers new",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				// findFirstNewHeader finds index 2
				mockLC.On("ContainsBlock", ctx, testBlocks[0].BlockHash()).Return(true, nil).Once()
				mockLC.On("ContainsBlock", ctx, testBlocks[1].BlockHash()).Return(true, nil).Once()
				mockLC.On("ContainsBlock", ctx, testBlocks[2].BlockHash()).Return(false, nil).Once()
				// then submitHeaderMessages calls InsertHeaders for chunks 2 and 3
				mockLC.On("InsertHeaders", ctx, expectedHeadersChunk2).Return(nil).Once() // 102, 103
				mockLC.On("InsertHeaders", ctx, expectedHeadersChunk3).Return(nil).Once() // 104
			},
			expectedCount: 3,
			expectedErr:   false,
		},
		{
			name: "no new headers",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				// findFirstNewHeader finds index -1
				for _, b := range testBlocks {
					mockLC.On("ContainsBlock", ctx, b.BlockHash()).Return(true, nil).Once()
				}
				// InsertHeaders should not be called
			},
			expectedCount: 0,
			expectedErr:   false,
		},
		{
			name: "error first chunk",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				// findFirstNewHeader finds index 0
				mockLC.On("ContainsBlock", ctx, testBlocks[0].BlockHash()).Return(false, nil).Once()
				// then submitHeaderMessages calls InsertHeaders and fails for chunk 1
				submitErr := fmt.Errorf("%w: ... MoveAbort(...)", sui_errors.ErrSuiTransactionFailed)
				mockLC.On("InsertHeaders", ctx, expectedHeadersChunk1).Return(submitErr).Once()
				// InsertHeaders should not be called after err
			},
			expectedCount: 0,
			expectedErr:   true,
		},
		{
			name: "error second chunk",
			mockSetup: func(mockLC *mocks.MockBitcoinSPV) {
				// findFirstNewHeader finds index 0
				mockLC.On("ContainsBlock", ctx, testBlocks[0].BlockHash()).Return(false, nil).Once()
				// then succeeds for first chunk
				mockLC.On("InsertHeaders", ctx, expectedHeadersChunk1).Return(nil).Once()
				// then fails for second chunk
				submitErr := fmt.Errorf("%w: ... MoveAbort(...)", sui_errors.ErrSuiTransactionFailed)
				mockLC.On("InsertHeaders", ctx, expectedHeadersChunk2).Return(submitErr).Once()
				// InsertHeaders should not be called after err
			},
			expectedCount: 0,
			expectedErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLC := mocks.NewMockBitcoinSPV(t)
			tt.mockSetup(mockLC)

			r := &Relayer{
				lcClient: mockLC,
				logger:   zerolog.Nop(),
				Config:   testSubmitConfig,
			}

			count, err := r.ProcessHeaders(ctx, testBlocks)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Equal(t, 0, count, "Count should be 0 when error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}
			mockLC.AssertExpectations(t)
		})
	}
}
