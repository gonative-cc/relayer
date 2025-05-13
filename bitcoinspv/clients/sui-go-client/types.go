package suigoclient

import (
	"bytes"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// BlockHeader is block header
type BlockHeader struct {
	Internal []uint8
}

// LightBlock is light block
type LightBlock struct {
	Height    uint64
	ChainWork [32]uint8
	Header    *BlockHeader
}

// BlockHash return block hash
func (lb LightBlock) BlockHash() chainhash.Hash {
	r := bytes.NewReader(lb.Header.Internal)
	var header wire.BlockHeader
	header.Deserialize(r)
	return header.BlockHash()
}
