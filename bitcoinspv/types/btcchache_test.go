package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBTCCache(t *testing.T) {
	t.Run("valid size", func(t *testing.T) {
		cache, err := NewBTCCache(10)
		assert.NoError(t, err)
		assert.NotNil(t, cache)
		assert.Equal(t, int64(10), cache.maxEntries)
		assert.Equal(t, int64(0), cache.Size())
		assert.NotNil(t, cache.blocks)
		assert.Equal(t, 0, len(cache.blocks))
	})

	t.Run("invalid size zero", func(t *testing.T) {
		cache, err := NewBTCCache(0)
		assert.ErrorIs(t, err, errCacheIncorrectMaxEntries)
		assert.Nil(t, cache)
	})

	t.Run("invalid size negative", func(t *testing.T) {
		cache, err := NewBTCCache(-1)
		assert.ErrorIs(t, err, errCacheIncorrectMaxEntries)
		assert.Nil(t, cache)
	})
}

func TestBTCCache_Init(t *testing.T) {
	maxSize := int64(5)

	t.Run("empty input", func(t *testing.T) {
		cache, _ := NewBTCCache(maxSize)
		blocks := CreateTestIndexedBlocks(t, 0, 0)
		err := cache.Init(blocks)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), cache.Size())
	})

	t.Run("nil input", func(t *testing.T) {
		cache, _ := NewBTCCache(maxSize)
		err := cache.Init(nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), cache.Size())
	})

	t.Run("valid input less than max", func(t *testing.T) {
		cache, _ := NewBTCCache(maxSize)
		blocks := CreateTestIndexedBlocks(t, 3, 100) // 100, 101, 102
		err := cache.Init(blocks)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), cache.Size())
		assert.Equal(t, int64(100), cache.First().BlockHeight)
		assert.Equal(t, int64(102), cache.Last().BlockHeight)
	})

	t.Run("valid input equal to max", func(t *testing.T) {
		cache, _ := NewBTCCache(maxSize)
		blocks := CreateTestIndexedBlocks(t, 5, 100) // 100, 101, 102, 103, 104
		err := cache.Init(blocks)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), cache.Size())
		assert.Equal(t, int64(100), cache.First().BlockHeight)
		assert.Equal(t, int64(104), cache.Last().BlockHeight)
	})

	t.Run("input exceeds max", func(t *testing.T) {
		cache, _ := NewBTCCache(maxSize)
		blocks := CreateTestIndexedBlocks(t, 6, 100) // 100, 101, 102, 103, 104, 105
		err := cache.Init(blocks)
		assert.ErrorIs(t, err, errBlockEntriesExceeded)
		assert.Equal(t, int64(0), cache.Size())
	})

	t.Run("unsorted input", func(t *testing.T) {
		cache, _ := NewBTCCache(maxSize)
		blocks := CreateTestIndexedBlocks(t, 3, 100) // 100, 101, 102
		blocks[0], blocks[1] = blocks[1], blocks[0]
		err := cache.Init(blocks)
		assert.ErrorIs(t, err, errUnorderedBlocks)
		assert.Equal(t, int64(0), cache.Size())
	})

	t.Run("re-init cache", func(t *testing.T) {
		cache, _ := NewBTCCache(maxSize)
		initialBlocks := CreateTestIndexedBlocks(t, 2, 50) // 50, 51
		_ = cache.Init(initialBlocks)
		assert.Equal(t, int64(2), cache.Size())

		newBlocks := CreateTestIndexedBlocks(t, 3, 100) // 100, 101, 102
		err := cache.Init(newBlocks)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), cache.Size())
		assert.Equal(t, int64(100), cache.First().BlockHeight)
		assert.Equal(t, int64(102), cache.Last().BlockHeight)
	})
}

func TestBTCCache_Add(t *testing.T) {
	maxSize := int64(3)
	cache, _ := NewBTCCache(maxSize)
	blocks := CreateTestIndexedBlocks(t, 5, 100) // 100, 101, 102, 103, 104

	cache.Add(blocks[0])
	assert.Equal(t, int64(1), cache.Size())
	assert.Equal(t, int64(100), cache.First().BlockHeight)
	assert.Equal(t, int64(100), cache.Last().BlockHeight)

	cache.Add(blocks[1])
	assert.Equal(t, int64(2), cache.Size())
	assert.Equal(t, int64(100), cache.First().BlockHeight)
	assert.Equal(t, int64(101), cache.Last().BlockHeight)

	// cache full
	cache.Add(blocks[2])
	assert.Equal(t, int64(3), cache.Size())
	assert.Equal(t, int64(100), cache.First().BlockHeight)
	assert.Equal(t, int64(102), cache.Last().BlockHeight)

	// should remove (100) oldest block.
	cache.Add(blocks[3]) // Add 103
	assert.Equal(t, int64(3), cache.Size())
	assert.Equal(t, int64(101), cache.First().BlockHeight)
	assert.Equal(t, int64(103), cache.Last().BlockHeight)

	// should remove (101) oldest block.
	cache.Add(blocks[4]) // Add 104
	assert.Equal(t, int64(3), cache.Size())
	assert.Equal(t, int64(102), cache.First().BlockHeight)
	assert.Equal(t, int64(104), cache.Last().BlockHeight)
}

func TestBTCCache_First_Last(t *testing.T) {
	maxSize := int64(3)
	cache, _ := NewBTCCache(maxSize)
	blocks := CreateTestIndexedBlocks(t, 3, 100) // 100, 101, 102

	assert.Equal(t, int64(0), cache.Size())
	assert.Nil(t, cache.First())
	assert.Nil(t, cache.Last())

	cache.Add(blocks[0])
	assert.Equal(t, int64(1), cache.Size())
	assert.Equal(t, blocks[0], cache.First())
	assert.Equal(t, blocks[0], cache.Last())
	assert.Equal(t, int64(100), cache.First().BlockHeight)

	cache.Add(blocks[1])
	cache.Add(blocks[2])
	assert.Equal(t, int64(3), cache.Size())
	assert.Equal(t, blocks[0], cache.First())
	assert.Equal(t, blocks[2], cache.Last())
	assert.Equal(t, int64(100), cache.First().BlockHeight)
	assert.Equal(t, int64(102), cache.Last().BlockHeight)
}

func TestBTCCache_RemoveLast(t *testing.T) {
	maxSize := int64(3)
	cache, _ := NewBTCCache(maxSize)
	blocks := CreateTestIndexedBlocks(t, 3, 100) // 100, 101, 102
	cache.Init(blocks)

	// 102
	err := cache.RemoveLast()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), cache.Size())
	assert.Equal(t, int64(101), cache.Last().BlockHeight)
	assert.Equal(t, int64(100), cache.First().BlockHeight)

	// 101
	err = cache.RemoveLast()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), cache.Size())
	assert.Equal(t, int64(100), cache.Last().BlockHeight)
	assert.Equal(t, int64(100), cache.First().BlockHeight)

	// 100
	err = cache.RemoveLast()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), cache.Size())
	assert.Nil(t, cache.Last())
	assert.Nil(t, cache.First())

	// try remove from empty
	err = cache.RemoveLast()
	assert.ErrorIs(t, err, errEmptyBlockCache)
	assert.Equal(t, int64(0), cache.Size())
}

func TestBTCCache_RemoveAll(t *testing.T) {
	cache, _ := NewBTCCache(5)

	// try remove from empty
	cache.RemoveAll()
	assert.Equal(t, int64(0), cache.Size())

	blocks := CreateTestIndexedBlocks(t, 3, 100)
	cache.Init(blocks)
	assert.Equal(t, int64(3), cache.Size())
	cache.RemoveAll()
	assert.Equal(t, int64(0), cache.Size())
	assert.Nil(t, cache.First())
	assert.Nil(t, cache.Last())
	assert.Empty(t, cache.blocks)
}

func TestBTCCache_GetAllBlocks(t *testing.T) {
	cache, _ := NewBTCCache(5)

	// empty
	all := cache.GetAllBlocks()
	assert.Empty(t, all)

	// populated
	blocks := CreateTestIndexedBlocks(t, 3, 100)
	cache.Init(blocks)
	all = cache.GetAllBlocks()
	assert.Len(t, all, 3)
	assert.Equal(t, int64(100), all[0].BlockHeight)
	assert.Equal(t, int64(102), all[2].BlockHeight)
}

func TestBTCCache_TrimConfirmedBlocks(t *testing.T) {
	blocks := CreateTestIndexedBlocks(t, 5, 100) // 100, 101, 102, 103, 104

	tests := []struct {
		name                string
		initialSize         int
		keepNMostRecent     int
		expectedTrimmedLen  int
		expectedFinalSize   int64
		expectedFirstHeight int64
		expectedLastHeight  int64
	}{
		{
			name:                "k less than size",
			initialSize:         5,
			keepNMostRecent:     3,
			expectedTrimmedLen:  2,
			expectedFinalSize:   3,
			expectedFirstHeight: 102,
			expectedLastHeight:  104,
		},
		{
			name:                "k equal to size",
			initialSize:         5,
			keepNMostRecent:     5,
			expectedTrimmedLen:  0,
			expectedFinalSize:   5,
			expectedFirstHeight: 100,
			expectedLastHeight:  104,
		},
		{
			name:                "k greater than size",
			initialSize:         5,
			keepNMostRecent:     6,
			expectedTrimmedLen:  0,
			expectedFinalSize:   5,
			expectedFirstHeight: 100,
			expectedLastHeight:  104,
		},
		{
			name:               "k is zero",
			initialSize:        5,
			keepNMostRecent:    0,
			expectedTrimmedLen: 5,
			expectedFinalSize:  0,
		},
		{
			name:                "k is one",
			initialSize:         5,
			keepNMostRecent:     1,
			expectedTrimmedLen:  4,
			expectedFinalSize:   1,
			expectedFirstHeight: 104,
			expectedLastHeight:  104,
		},
		{
			name:               "initial size zero",
			initialSize:        0,
			keepNMostRecent:    3,
			expectedTrimmedLen: 0,
			expectedFinalSize:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, _ := NewBTCCache(10)
			if tt.initialSize > 0 {
				cache.Init(blocks[:tt.initialSize])
			}

			trimmed := cache.TrimConfirmedBlocks(tt.keepNMostRecent)

			if tt.expectedTrimmedLen == 0 {
				assert.Nil(t, trimmed)
			} else {
				assert.Len(t, trimmed, tt.expectedTrimmedLen)
			}

			assert.Equal(t, tt.expectedFinalSize, cache.Size())
			if tt.expectedFinalSize > 0 {
				assert.Equal(t, tt.expectedFirstHeight, cache.First().BlockHeight)
				assert.Equal(t, tt.expectedLastHeight, cache.Last().BlockHeight)
			} else {
				assert.Nil(t, cache.First())
				assert.Nil(t, cache.Last())
			}
		})
	}
}

func TestBTCCache_FindBlock(t *testing.T) {
	cache, _ := NewBTCCache(10)
	blocks := CreateTestIndexedBlocks(t, 5, 100) // 100, 101, 102, 103, 104
	cache.Init(blocks)

	tests := []struct {
		name   string
		height int64
		err    error
	}{
		{name: "find first", height: 100, err: nil},
		{name: "find middle", height: 102, err: nil},
		{name: "find last", height: 104, err: nil},
		{name: "find miss 1", height: 99, err: errors.New("block at height 99 not found")},
		{name: "find miss 2", height: 105, err: errors.New("block at height 105 not found")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundBlock, err := cache.FindBlock(tt.height)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.height, foundBlock.BlockHeight)
			}
		})
	}

	t.Run("find in empty cache", func(t *testing.T) {
		emptyCache, _ := NewBTCCache(5)
		_, err := emptyCache.FindBlock(100)
		assert.EqualError(t, err, "cache is empty")
	})
}

func TestBTCCache_Resize(t *testing.T) {
	cache, _ := NewBTCCache(5)

	err := cache.Resize(10)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), cache.maxEntries)

	err = cache.Resize(3)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), cache.maxEntries)

	err = cache.Resize(0)
	assert.ErrorIs(t, err, errCacheIncorrectMaxEntries)
	assert.Equal(t, int64(3), cache.maxEntries)

	err = cache.Resize(-2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errCacheIncorrectMaxEntries)
	assert.Equal(t, int64(3), cache.maxEntries)
}

func TestBTCCache_Trim(t *testing.T) {
	t.Run("no trim", func(t *testing.T) {
		cache, _ := NewBTCCache(5)
		cache.Init(CreateTestIndexedBlocks(t, 3, 100))
		cache.Trim()
		assert.Equal(t, int64(3), cache.Size())
		assert.Equal(t, int64(100), cache.First().BlockHeight)
	})

	t.Run("trim", func(t *testing.T) {
		cache, _ := NewBTCCache(5)
		cache.Init(CreateTestIndexedBlocks(t, 5, 100)) // 100, 101, 102, 103, 104

		err := cache.Resize(3)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), cache.Size())

		cache.Trim() // should remove 100, 101

		assert.Equal(t, int64(3), cache.Size())
		assert.Equal(t, int64(102), cache.First().BlockHeight)
		assert.Equal(t, int64(104), cache.Last().BlockHeight)
	})

	t.Run("trim already correct size", func(t *testing.T) {
		cache, _ := NewBTCCache(3)
		cache.Init(CreateTestIndexedBlocks(t, 3, 100))
		cache.Trim()
		assert.Equal(t, int64(3), cache.Size())
		assert.Equal(t, int64(100), cache.First().BlockHeight)
	})

	t.Run("trim empty cache", func(t *testing.T) {
		cache, _ := NewBTCCache(3)
		cache.Trim()
		assert.Equal(t, int64(0), cache.Size())
	})
}
