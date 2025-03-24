package types

import "errors"

var (
	errEmptyBlockCache          = errors.New("empty block cache")
	errCacheIncorrectMaxEntries = errors.New("incorrect max entries")
	errBlockEntriesExceeded     = errors.New("number of blocks is more than maxEntries")
	errUnorderedBlocks          = errors.New("blocks are not sorted by height")
	errIndexOutOfBounds         = errors.New("provided index should be smaller that length of transaction list")
	errEmptyTxList              = errors.New("can't calculate proof for empty transaction list")
)
