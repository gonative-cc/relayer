package bitcoinspv

import (
	"context"
	"testing"

	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/mocks"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBreakIntoChunks(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		chunkSize int
		want      [][]int
	}{
		{
			name:      "empty slice",
			input:     []int{},
			chunkSize: 2,
			want:      nil,
		},
		{
			name:      "single chunk",
			input:     []int{1, 2, 3},
			chunkSize: 3,
			want:      [][]int{{1, 2, 3}},
		},
		{
			name:      "multiple chunks",
			input:     []int{1, 2, 3, 4, 5},
			chunkSize: 2,
			want:      [][]int{{1, 2}, {3, 4}, {5}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := breakIntoChunks(tt.input, tt.chunkSize)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetHeaderMessages(t *testing.T) {
	ctx := context.Background()
	mockLC := mocks.NewBitcoinSPV(t)

	// Create test blocks
	block1 := &types.IndexedBlock{
		BlockHeight: 1,
		BlockHeader: &wire.BlockHeader{},
	}
	block2 := &types.IndexedBlock{
		BlockHeight: 2,
		BlockHeader: &wire.BlockHeader{},
	}
	block3 := &types.IndexedBlock{
		BlockHeight: 3,
		BlockHeader: &wire.BlockHeader{},
	}

	tests := []struct {
		name          string
		indexedBlocks []*types.IndexedBlock
		mockSetup     func()
		wantHeaders   int
		wantErr       bool
	}{
		{
			name:          "all headers are new",
			indexedBlocks: []*types.IndexedBlock{block1, block2, block3},
			mockSetup: func() {
				mockLC.On("ContainsBlock", ctx, block1.BlockHash()).Return(false, nil)
				mockLC.On("ContainsBlock", ctx, block2.BlockHash()).Return(false, nil)
				mockLC.On("ContainsBlock", ctx, block3.BlockHash()).Return(false, nil)
			},
			wantHeaders: 3,
			wantErr:     false,
		},
		{
			name:          "some headers are new",
			indexedBlocks: []*types.IndexedBlock{block1, block2, block3},
			mockSetup: func() {
				mockLC.On("ContainsBlock", ctx, block1.BlockHash()).Return(true, nil)
				mockLC.On("ContainsBlock", ctx, block2.BlockHash()).Return(false, nil)
				mockLC.On("ContainsBlock", ctx, block3.BlockHash()).Return(false, nil)
			},
			wantHeaders: 2,
			wantErr:     false,
		},
		{
			name:          "all headers exist",
			indexedBlocks: []*types.IndexedBlock{block1, block2, block3},
			mockSetup: func() {
				mockLC.On("ContainsBlock", ctx, block1.BlockHash()).Return(true, nil)
				mockLC.On("ContainsBlock", ctx, block2.BlockHash()).Return(true, nil)
				mockLC.On("ContainsBlock", ctx, block3.BlockHash()).Return(true, nil)
			},
			wantHeaders: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Relayer{
				lcClient: mockLC,
			}

			tt.mockSetup()

			headers, err := r.getHeaderMessages(ctx, tt.indexedBlocks)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.wantHeaders == 0 {
				assert.Nil(t, headers)
				return
			}

			totalHeaders := 0
			for _, chunk := range headers {
				totalHeaders += len(chunk)
			}
			assert.Equal(t, tt.wantHeaders, totalHeaders)
		})
	}
}

func TestSubmitHeaderMessages(t *testing.T) {
	ctx := context.Background()
	mockLC := mocks.NewBitcoinSPV(t)

	headers := []wire.BlockHeader{
		{Version: 1},
		{Version: 2},
	}

	tests := []struct {
		name      string
		headers   []wire.BlockHeader
		mockSetup func()
		wantErr   bool
	}{
		{
			name:    "successful submission",
			headers: headers,
			mockSetup: func() {
				mockLC.On("InsertHeaders", ctx, headers).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "failed submission",
			headers: headers,
			mockSetup: func() {
				mockLC.On("InsertHeaders", ctx, headers).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Relayer{
				lcClient: mockLC,
			}

			tt.mockSetup()

			err := r.submitHeaderMessages(ctx, tt.headers)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			mockLC.AssertExpectations(t)
		})
	}
}

func TestProcessHeaders(t *testing.T) {
	ctx := context.Background()
	mockLC := mocks.NewBitcoinSPV(t)

	// Create test blocks
	block1 := &types.IndexedBlock{
		BlockHeight: 1,
		BlockHeader: &wire.BlockHeader{},
	}
	block2 := &types.IndexedBlock{
		BlockHeight: 2,
		BlockHeader: &wire.BlockHeader{},
	}

	tests := []struct {
		name          string
		indexedBlocks []*types.IndexedBlock
		mockSetup     func()
		wantCount     int
		wantErr       bool
	}{
		{
			name:          "successful processing",
			indexedBlocks: []*types.IndexedBlock{block1, block2},
			mockSetup: func() {
				mockLC.On("ContainsBlock", ctx, block1.BlockHash()).Return(false, nil)
				mockLC.On("ContainsBlock", ctx, block2.BlockHash()).Return(false, nil)
				mockLC.On("InsertHeaders", ctx, mock.Anything).Return(nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:          "no new headers",
			indexedBlocks: []*types.IndexedBlock{block1, block2},
			mockSetup: func() {
				mockLC.On("ContainsBlock", ctx, block1.BlockHash()).Return(true, nil)
				mockLC.On("ContainsBlock", ctx, block2.BlockHash()).Return(true, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Relayer{
				lcClient: mockLC,
			}

			tt.mockSetup()

			count, err := r.ProcessHeaders(ctx, tt.indexedBlocks)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
			mockLC.AssertExpectations(t)
		})
	}
}
