package types

import (
	"sort"
	"sync"
)

// BTCCache represents a thread-safe cache of indexed blocks
//
//nolint:govet
type BTCCache struct {
	sync.RWMutex
	blocks     []*IndexedBlock
	maxEntries uint64
}

// NewBTCCache creates a new BTCCache with the specified max entries
func NewBTCCache(maxEntries uint64) (*BTCCache, error) {
	if maxEntries <= 0 {
		return nil, errCacheIncorrectMaxEntries
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

	if uint64(len(blocks)) > cache.maxEntries {
		return errBlockEntriesExceeded
	}

	if !sort.SliceIsSorted(blocks, func(i, j int) bool {
		return blocks[i].BlockHeight < blocks[j].BlockHeight
	}) {
		return errUnorderedBlocks
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

// TODO: return the error rather than panicing.
func (cache *BTCCache) add(block *IndexedBlock) {
	if cache.size() > cache.maxEntries {
		panic(errBlockEntriesExceeded)
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
		return errEmptyBlockCache
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
func (cache *BTCCache) Size() uint64 {
	// TODO: do we need mutex just for read?
	cache.RLock()
	defer cache.RUnlock()
	return cache.size()
}

func (cache *BTCCache) size() uint64 {
	return uint64(len(cache.blocks))
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

		switch {
		case blockHeight == height:
			return block
		case blockHeight > height:
			right = mid - 1
		default:
			left = mid + 1
		}
	}

	return nil
}

// TODO: lets merge those two functions and call it resize_and_trim

// Resize updates the maximum number of entries allowed in the cache
func (cache *BTCCache) Resize(maxEntries uint64) error {
	cache.Lock()
	defer cache.Unlock()

	if maxEntries <= 0 {
		return errCacheIncorrectMaxEntries
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

	trimAt := uint64(len(cache.blocks)) - cache.maxEntries

	for i := range cache.blocks[:trimAt] {
		cache.blocks[i] = nil
	}

	cache.blocks = cache.blocks[trimAt:]
}
