package bitcoinspv

import (
	"testing"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/stretchr/testify/assert"
)

func TestToBlockHeaders(t *testing.T) {
	t.Run("single block", func(t *testing.T) {
		inputBlocks := types.CreateTestIndexedBlocks(t, 1, 100)
		headers := toBlockHeaders(inputBlocks)
		assert.NotNil(t, headers)
		assert.Len(t, headers, 1)
		assert.Equal(t, inputBlocks[0].BlockHeader.Version, headers[0].Version)
	})

	t.Run("multiple blocks", func(t *testing.T) {
		inputBlocks := types.CreateTestIndexedBlocks(t, 3, 200)
		headers := toBlockHeaders(inputBlocks)
		assert.NotNil(t, headers)
		assert.Len(t, headers, 3)
		for i := 0; i < 3; i++ {
			assert.Equal(t, inputBlocks[i].BlockHeader.Version, headers[i].Version)
		}
	})
}

func TestBreakIntoChunks(t *testing.T) {
	tests := []struct {
		name              string
		inputBlocks       []*types.IndexedBlock
		chunkSize         int
		expectedNumChunks int
		expectedFromTo    [][]int64 // [From, To] for each chunk
	}{
		{
			name:              "empty slice",
			inputBlocks:       types.CreateTestIndexedBlocks(t, 0, 0),
			chunkSize:         2,
			expectedNumChunks: 0,
			expectedFromTo:    nil,
		},
		{
			name:              "nil slice",
			inputBlocks:       nil,
			chunkSize:         2,
			expectedNumChunks: 0,
			expectedFromTo:    nil,
		},
		{
			name:              "zero chunk size",
			inputBlocks:       types.CreateTestIndexedBlocks(t, 5, 100),
			chunkSize:         0,
			expectedNumChunks: 0,
			expectedFromTo:    nil,
		},
		{
			name:              "negative chunk size",
			inputBlocks:       types.CreateTestIndexedBlocks(t, 5, 100),
			chunkSize:         -1,
			expectedNumChunks: 0,
			expectedFromTo:    nil,
		},
		{
			name:              "single chunk exact size",
			inputBlocks:       types.CreateTestIndexedBlocks(t, 3, 100), // Heights 100, 101, 102
			chunkSize:         3,
			expectedNumChunks: 1,
			expectedFromTo:    [][]int64{{100, 102}},
		},
		{
			name:              "single chunk smaller than size",
			inputBlocks:       types.CreateTestIndexedBlocks(t, 2, 100), // Heights 100, 101
			chunkSize:         3,
			expectedNumChunks: 1,
			expectedFromTo:    [][]int64{{100, 101}},
		},
		{
			name:              "multiple full chunks",
			inputBlocks:       types.CreateTestIndexedBlocks(t, 4, 100), // Heights 100, 101, 102, 103
			chunkSize:         2,
			expectedNumChunks: 2,
			expectedFromTo:    [][]int64{{100, 101}, {102, 103}},
		},
		{
			name:              "multiple chunks partial last",
			inputBlocks:       types.CreateTestIndexedBlocks(t, 5, 100), // Heights 100, 101, 102, 103, 104
			chunkSize:         2,
			expectedNumChunks: 3,
			expectedFromTo:    [][]int64{{100, 101}, {102, 103}, {104, 104}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChunks := breakIntoChunks(tt.inputBlocks, tt.chunkSize)

			if tt.expectedNumChunks == 0 {
				assert.Nil(t, gotChunks, "Expected nil or empty slice for chunks")
				return
			}

			assert.NotNil(t, gotChunks)
			assert.Len(t, gotChunks, tt.expectedNumChunks)
			assert.Len(t, tt.expectedFromTo, tt.expectedNumChunks)

			for i, chunk := range gotChunks {
				assert.Equal(t, tt.expectedFromTo[i][0], chunk.From, "Chunk %d: Unexpected From height", i)
				assert.Equal(t, tt.expectedFromTo[i][1], chunk.To, "Chunk %d: Unexpected To height", i)
			}
		})
	}
}
