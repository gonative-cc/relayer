package btc

import (
	"bytes"
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

const (
	// HeaderLen is the fixed length of a Bitcoin block header in bytes.
	HeaderLen = 80
	// HeaderHashLen is length of hash of a BTC header
	HeaderHashLen = 32
)

// HeaderBytes represents a Bitcoin block header as a byte slice.
type HeaderBytes []byte

// HeaderHashBytes represents a Bitcoin header hash as a byte slice
type HeaderHashBytes []byte

// EventType represents different types of blockchain events
type EventType int

const (
	// BlockDisconnected is triggered when a block is removed from the chain
	BlockDisconnected EventType = iota
	// BlockConnected is triggered when a block is added to the chain
	BlockConnected
)

// BlockEvent contains information about a blockchain event
type BlockEvent struct {
	BlockHeader *wire.BlockHeader
	Type        EventType
	Height      int64
}

// NewBlockEvent creates a new block event with the given parameters
func NewBlockEvent(evtType EventType, blockHeight int64, blockHeader *wire.BlockHeader) *BlockEvent {
	evt := &BlockEvent{
		Type:        evtType,
		Height:      blockHeight,
		BlockHeader: blockHeader,
	}
	return evt
}

// NewHeaderBytesFromBlockHeader creates a new HeaderBytes from a wire.BlockHeader.
func NewHeaderBytesFromBlockHeader(header *wire.BlockHeader) HeaderBytes {
	headerBytes := HeaderBytes{}
	headerBytes.FromBlockHeader(header)
	return headerBytes
}

// Marshal returns the byte slice representation of HeaderBytes.
func (b HeaderBytes) Marshal() ([]byte, error) {
	return b, nil
}

// Unmarshal validates and sets HeaderBytes from a byte slice.
// Returns error if data length is not 80 bytes or if data is not a valid block header.
func (b *HeaderBytes) Unmarshal(data []byte) error {
	if len(data) != HeaderLen {
		return errors.New("header length must be exactly 80 bytes")
	}

	if _, err := NewBlockHeader(data); err != nil {
		return errors.New("failed to parse bytes as valid block header")
	}

	*b = data
	return nil
}

// Size returns the length of the HeaderBytes in bytes.
func (b *HeaderBytes) Size() int {
	bytes, _ := b.Marshal()
	return len(bytes)
}

// ToBlockHeader converts HeaderBytes to a wire.BlockHeader.
// Panics if conversion fails.
func (b HeaderBytes) ToBlockHeader() *wire.BlockHeader {
	header, err := NewBlockHeader(b)
	if err != nil {
		panic("failed to convert bytes to block header format")
	}
	return header
}

// FromBlockHeader converts a wire.BlockHeader to HeaderBytes.
// Panics if serialization or validation fails.
func (b *HeaderBytes) FromBlockHeader(header *wire.BlockHeader) {
	buffer := bytes.Buffer{}
	if err := header.Serialize(&buffer); err != nil {
		panic("failed to serialize block header to bytes")
	}

	if err := b.Unmarshal(buffer.Bytes()); err != nil {
		panic("failed to validate serialized block header bytes")
	}
}

// Eq checks if two HeaderBytes are equal by comparing their hashes.
func (b *HeaderBytes) Eq(other *HeaderBytes) bool {
	return b.Hash().Eq(other.Hash())
}

// Hash returns the hash of the block header as HeaderHashBytes.
func (b *HeaderBytes) Hash() *HeaderHashBytes {
	hash := b.ToBlockHeader().BlockHash()
	headerHash := NewHeaderHashBytesFromChainhash(&hash)
	return &headerHash
}

// NewBlockHeader creates a block header from bytes.
// Returns error if deserialization fails.
func NewBlockHeader(data []byte) (*wire.BlockHeader, error) {
	header := &wire.BlockHeader{}
	reader := bytes.NewReader(data)
	return header, header.Deserialize(reader)
}

// Marshal returns the byte slice representation of the header hash
func (b HeaderHashBytes) Marshal() ([]byte, error) {
	return b, nil
}

// Unmarshal validates and sets the header hash from a byte slice
func (b *HeaderHashBytes) Unmarshal(data []byte) error {
	if len(data) != HeaderHashLen {
		return errors.New("header hash must be exactly 32 bytes")
	}

	if _, err := toChainhash(data); err != nil {
		return errors.New("failed to convert bytes to chainhash.Hash format")
	}

	*b = data
	return nil
}

// Size returns the length of the header hash in bytes
func (b *HeaderHashBytes) Size() int {
	data, _ := b.Marshal()
	return len(data)
}

// NewHeaderHashBytesFromChainhash creates a new HeaderHashBytes from a chainhash.Hash
func NewHeaderHashBytesFromChainhash(hash *chainhash.Hash) HeaderHashBytes {
	result := HeaderHashBytes{}
	result.FromChainhash(hash)
	return result
}

// ToChainhash converts the HeaderHashBytes to a chainhash.Hash pointer
func (b HeaderHashBytes) ToChainhash() *chainhash.Hash {
	result, err := toChainhash(b)
	if err != nil {
		panic("failed to convert HeaderHashBytes to chainhash format")
	}
	return result
}

// FromChainhash sets the HeaderHashBytes from a chainhash.Hash
func (b *HeaderHashBytes) FromChainhash(hash *chainhash.Hash) {
	if err := b.Unmarshal(hash[:]); err != nil {
		panic("failed to convert chainhash.Hash bytes to HeaderHashBytes")
	}
}

// String returns the string representation of the header hash
func (b *HeaderHashBytes) String() string {
	return b.ToChainhash().String()
}

// Eq checks if two HeaderHashBytes are equal
func (b *HeaderHashBytes) Eq(other *HeaderHashBytes) bool {
	return b.String() == other.String()
}

// toChainhash converts a byte slice to a chainhash.Hash pointer
func toChainhash(data []byte) (*chainhash.Hash, error) {
	return chainhash.NewHash(data)
}
