// revive:disable:var-naming

package types

import "errors"

var (
	errEmptyBlockCache          = errors.New("empty block cache")
	errCacheIncorrectMaxEntries = errors.New("incorrect max entries")
	errBlockEntriesExceeded     = errors.New("number of blocks is more than maxEntries")
	errUnorderedBlocks          = errors.New("blocks are not sorted by height")
)
