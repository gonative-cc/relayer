package types

import (
	"bytes"
	"errors"

	"github.com/btcsuite/btcd/wire"
)

type BTCHeaderBytes []byte

const BTCHeaderLen = 80

func NewBTCHeaderBytesFromBlockHeader(header *wire.BlockHeader) BTCHeaderBytes {
	headerBytes := BTCHeaderBytes{}
	headerBytes.FromBlockHeader(header)
	return headerBytes
}

func (b BTCHeaderBytes) Marshal() ([]byte, error) {
	return b, nil
}

func (b *BTCHeaderBytes) Unmarshal(data []byte) error {
	if len(data) != BTCHeaderLen {
		return errors.New("header length must be exactly 80 bytes")
	}

	if _, err := NewBlockHeader(data); err != nil {
		return errors.New("failed to parse bytes as valid block header")
	}

	*b = data
	return nil
}

func (b *BTCHeaderBytes) Size() int {
	bytes, _ := b.Marshal()
	return len(bytes)
}

func (b BTCHeaderBytes) ToBlockHeader() *wire.BlockHeader {
	if header, err := NewBlockHeader(b); err != nil {
		panic("failed to convert bytes to block header format")
	} else {
		return header
	}
}

func (b *BTCHeaderBytes) FromBlockHeader(header *wire.BlockHeader) {
	buffer := bytes.Buffer{}
	if err := header.Serialize(&buffer); err != nil {
		panic("failed to serialize block header to bytes")
	}

	if err := b.Unmarshal(buffer.Bytes()); err != nil {
		panic("failed to validate serialized block header bytes")
	}
}

func (b *BTCHeaderBytes) Eq(other *BTCHeaderBytes) bool {
	return b.Hash().Eq(other.Hash())
}

func (b *BTCHeaderBytes) Hash() *BTCHeaderHashBytes {
	hash := b.ToBlockHeader().BlockHash()
	headerHash := NewBTCHeaderHashBytesFromChainhash(&hash)
	return &headerHash
}

// NewBlockHeader creates a block header from bytes.
func NewBlockHeader(data []byte) (*wire.BlockHeader, error) {
	header := &wire.BlockHeader{}
	reader := bytes.NewReader(data)
	return header, header.Deserialize(reader)
}
