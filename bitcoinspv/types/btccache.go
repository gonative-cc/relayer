package types

import (
	"fmt"
	"sort"
	"sync"
)

type BTCCache struct {
	blocks     []*IndexedBlock
	maxEntries int64
	sync.RWMutex
}

func NewBTCCache(maxEntries int64) (*BTCCache, error) {
	if maxEntries == 0 {
		return nil, ErrInvalidMaxEntries
	}

	return &BTCCache{
		blocks:     make([]*IndexedBlock, 0, maxEntries),
		maxEntries: maxEntries,
	}, nil
}

func (b *BTCCache) Init(ibs []*IndexedBlock) error {
	b.Lock()
	defer b.Unlock()

	if int64(len(ibs)) > b.maxEntries {
		return ErrTooManyEntries
	}

	if !sort.SliceIsSorted(ibs, func(i, j int) bool {
		return ibs[i].Height < ibs[j].Height
	}) {
		return ErrorUnsortedBlocks
	}

	for _, ib := range ibs {
		b.add(ib)
	}

	return nil
}

func (b *BTCCache) Add(ib *IndexedBlock) {
	b.Lock()
	defer b.Unlock()
	b.add(ib)
}

func (b *BTCCache) add(ib *IndexedBlock) {
	if b.size() > b.maxEntries {
		panic(ErrTooManyEntries)
	}

	if b.size() == b.maxEntries {
		// dereference the 0-th block to ensure it will be garbage-collected
		// see https://stackoverflow.com/questions/55045402/memory-leak-in-golang-slice
		b.blocks[0] = nil
		b.blocks = b.blocks[1:]
	}

	b.blocks = append(b.blocks, ib)
}

func (b *BTCCache) First() *IndexedBlock {
	b.RLock()
	defer b.RUnlock()

	if b.size() == 0 {
		return nil
	}
	return b.blocks[0]
}

func (b *BTCCache) Tip() *IndexedBlock {
	b.RLock()
	defer b.RUnlock()

	if b.size() == 0 {
		return nil
	}
	return b.blocks[len(b.blocks)-1]
}

func (b *BTCCache) RemoveLast() error {
	b.Lock()
	defer b.Unlock()

	if b.size() == 0 {
		return ErrEmptyCache
	}

	lastIdx := len(b.blocks) - 1
	b.blocks[lastIdx] = nil
	b.blocks = b.blocks[:lastIdx]
	return nil
}

func (b *BTCCache) RemoveAll() {
	b.Lock()
	defer b.Unlock()
	b.blocks = []*IndexedBlock{}
}

func (b *BTCCache) Size() int64 {
	b.RLock()
	defer b.RUnlock()
	return b.size()
}

func (b *BTCCache) size() int64 {
	return int64(len(b.blocks))
}

func (b *BTCCache) GetLastBlocks(stopHeight int64) ([]*IndexedBlock, error) {
	b.RLock()
	defer b.RUnlock()

	if len(b.blocks) == 0 {
		return []*IndexedBlock{}, nil
	}

	firstHeight := b.blocks[0].Height
	lastHeight := b.blocks[len(b.blocks)-1].Height

	if stopHeight < firstHeight || lastHeight < stopHeight {
		return []*IndexedBlock{}, fmt.Errorf(
			"the given stopHeight %d is out of range [%d, %d] of BTC cache",
			stopHeight, firstHeight, lastHeight,
		)
	}

	var startIdx int
	for i := len(b.blocks) - 1; i >= 0; i-- {
		if b.blocks[i].Height == stopHeight {
			startIdx = i
			break
		}
	}

	return b.blocks[startIdx:], nil
}

func (b *BTCCache) GetAllBlocks() []*IndexedBlock {
	b.RLock()
	defer b.RUnlock()
	return b.blocks
}

func (b *BTCCache) TrimConfirmedBlocks(k int) []*IndexedBlock {
	b.Lock()
	defer b.Unlock()

	l := len(b.blocks)
	if l <= k {
		return nil
	}

	trimmedBlocks := make([]*IndexedBlock, l-k)
	copy(trimmedBlocks, b.blocks)
	b.blocks = b.blocks[l-k:]

	return trimmedBlocks
}

// FindBlock uses binary search to find the block with the given height in cache
func (b *BTCCache) FindBlock(blockHeight int64) *IndexedBlock {
	b.RLock()
	defer b.RUnlock()

	if b.size() == 0 {
		return nil
	}

	firstHeight := b.blocks[0].Height
	lastHeight := b.blocks[len(b.blocks)-1].Height
	if blockHeight < firstHeight || lastHeight < blockHeight {
		return nil
	}

	left := 0
	right := len(b.blocks) - 1

	for left <= right {
		mid := left + (right-left)/2
		block := b.blocks[mid]

		switch {
		case block.Height == blockHeight:
			return block
		case block.Height > blockHeight:
			right = mid - 1
		default:
			left = mid + 1
		}
	}

	return nil
}

func (b *BTCCache) Resize(maxEntries int64) error {
	b.Lock()
	defer b.Unlock()

	if maxEntries == 0 {
		return ErrInvalidMaxEntries
	}
	b.maxEntries = maxEntries
	return nil
}

// Trim trims BTCCache to only keep the latest `maxEntries` blocks,
// and set `maxEntries` to be the cache size
func (b *BTCCache) Trim() {
	b.Lock()
	defer b.Unlock()

	// cache size is smaller than maxEntries, can't trim
	if b.size() < b.maxEntries {
		return
	}

	trimIndex := int64(len(b.blocks)) - b.maxEntries

	// dereference old blocks to ensure they will be garbage-collected
	for i := range b.blocks[:trimIndex] {
		b.blocks[i] = nil
	}

	b.blocks = b.blocks[trimIndex:]
}
