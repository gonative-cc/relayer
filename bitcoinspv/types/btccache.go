package types

import (
	"fmt"
	"sort"
	"sync"
)

// BTCCache represents a thread-safe cache of indexed blocks
type BTCCache struct {
	sync.RWMutex
	maxEntries int64
	blocks     []*IndexedBlock
}

// NewBTCCache creates a new BTCCache with the specified max entries
func NewBTCCache(maxEntries int64) (*BTCCache, error) {
	if maxEntries <= 0 {
		return nil, ErrInvalidMaxEntries
	}
	cache := &BTCCache{
		blocks:     make([]*IndexedBlock, 0, maxEntries),
		maxEntries: maxEntries,
	}
	return cache, nil
}

// Init initializes the cache with a slice of indexed blocks
func (cache *BTCCache) Init(blocks []*IndexedBlock) error {
	cache.Lock()
	defer cache.Unlock()

	if int64(len(blocks)) > cache.maxEntries {
		return ErrTooManyEntries
	}

	if !sort.SliceIsSorted(blocks, func(i, j int) bool {
		return blocks[i].BlockHeight < blocks[j].BlockHeight
	}) {
		return ErrorUnsortedBlocks
	}

	for _, block := range blocks {
		cache.add(block)
	}

	return nil
}

// Add adds a new block to the cache
func (cache *BTCCache) Add(block *IndexedBlock) {
	cache.Lock()
	defer cache.Unlock()
	cache.add(block)
}

func (cache *BTCCache) add(block *IndexedBlock) {
	if cache.size() > cache.maxEntries {
		panic(ErrTooManyEntries)
	}

	if cache.size() == cache.maxEntries {
		cache.blocks[0] = nil
		cache.blocks = cache.blocks[1:]
	}

	cache.blocks = append(cache.blocks, block)
}

// First returns the first block in the cache
func (cache *BTCCache) First() *IndexedBlock {
	cache.RLock()
	defer cache.RUnlock()

	if len(cache.blocks) == 0 {
		return nil
	}
	return cache.blocks[0]
}

// Tip returns the most recent block in the cache
func (cache *BTCCache) Tip() *IndexedBlock {
	cache.RLock()
	defer cache.RUnlock()

	if len(cache.blocks) == 0 {
		return nil
	}
	return cache.blocks[len(cache.blocks)-1]
}

// RemoveLast removes the most recent block from the cache
func (cache *BTCCache) RemoveLast() error {
	cache.Lock()
	defer cache.Unlock()

	if cache.size() == 0 {
		return ErrEmptyCache
	}

	lastIdx := len(cache.blocks) - 1
	cache.blocks[lastIdx] = nil
	cache.blocks = cache.blocks[:lastIdx]
	return nil
}

// RemoveAll removes all blocks from the cache
func (cache *BTCCache) RemoveAll() {
	cache.Lock()
	defer cache.Unlock()
	cache.blocks = make([]*IndexedBlock, 0)
}

// Size returns the current number of blocks in the cache
func (cache *BTCCache) Size() int64 {
	cache.RLock()
	defer cache.RUnlock()
	return cache.size()
}

func (cache *BTCCache) size() int64 {
	return int64(len(cache.blocks))
}

// GetLastBlocks returns blocks from the given height to the tip
func (cache *BTCCache) GetLastBlocks(height int64) ([]*IndexedBlock, error) {
	cache.RLock()
	defer cache.RUnlock()

	if len(cache.blocks) == 0 {
		return []*IndexedBlock{}, nil
	}

	first := cache.blocks[0].BlockHeight
	last := cache.blocks[len(cache.blocks)-1].BlockHeight

	if height < first || last < height {
		return nil, fmt.Errorf(
			"height %d is out of range [%d, %d] of BTC cache",
			height, first, last,
		)
	}

	// Use FindBlock to get the block at target height
	block := cache.FindBlock(height)
	if block == nil {
		return nil, fmt.Errorf("block at height %d not found", height)
	}

	// Find index of block
	idx := 0
	for i, b := range cache.blocks {
		if b == block {
			idx = i
			break
		}
	}

	return cache.blocks[idx:], nil
}

// GetAllBlocks returns all blocks in the cache
func (cache *BTCCache) GetAllBlocks() []*IndexedBlock {
	cache.RLock()
	defer cache.RUnlock()
	return cache.blocks
}

// TrimConfirmedBlocks removes confirmed blocks keeping only k most recent
func (cache *BTCCache) TrimConfirmedBlocks(k int) []*IndexedBlock {
	cache.Lock()
	defer cache.Unlock()

	size := len(cache.blocks)
	if size <= k {
		return nil
	}

	trimmed := make([]*IndexedBlock, size-k)
	copy(trimmed, cache.blocks)
	cache.blocks = cache.blocks[size-k:]

	return trimmed
}

// FindBlock locates a block by its height using binary search
func (cache *BTCCache) FindBlock(height int64) *IndexedBlock {
	cache.RLock()
	defer cache.RUnlock()

	blocks := cache.blocks
	if len(blocks) == 0 {
		return nil
	}

	// Check if height is within valid range
	if height < blocks[0].BlockHeight || height > blocks[len(blocks)-1].BlockHeight {
		return nil
	}

	// Binary search
	left, right := 0, len(blocks)-1
	for left <= right {
		mid := left + (right-left)/2
		block := blocks[mid]
		blockHeight := block.BlockHeight

		if blockHeight == height {
			return block
		} else if blockHeight > height {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}

	return nil
}

// Resize updates the maximum number of entries allowed in the cache
func (cache *BTCCache) Resize(maxEntries int64) error {
	cache.Lock()
	defer cache.Unlock()

	if maxEntries <= 0 {
		return ErrInvalidMaxEntries
	}
	cache.maxEntries = maxEntries
	return nil
}

// Trim removes oldest blocks to keep cache size within maxEntries limit
func (cache *BTCCache) Trim() {
	cache.Lock()
	defer cache.Unlock()

	if cache.size() <= cache.maxEntries {
		return
	}

	trimAt := int64(len(cache.blocks)) - cache.maxEntries

	for i := range cache.blocks[:trimAt] {
		cache.blocks[i] = nil
	}

	cache.blocks = cache.blocks[trimAt:]
}
