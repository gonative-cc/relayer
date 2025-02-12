package types

import (
	"bytes"
	"errors"

	"github.com/btcsuite/btcd/wire"
)

type BTCHeaderBytes []byte

const BTCHeaderLen = 80

func NewBTCHeaderBytesFromBlockHeader(header *wire.BlockHeader) BTCHeaderBytes {
	var headerBytes BTCHeaderBytes
	headerBytes.FromBlockHeader(header)
	return headerBytes
}

func (m BTCHeaderBytes) Marshal() ([]byte, error) {
	return m, nil
}

func (m *BTCHeaderBytes) Unmarshal(data []byte) error {
	if len(data) != BTCHeaderLen {
		return errors.New("invalid header length")
	}
	// Verify that the bytes can be transformed to a *wire.BlockHeader object
	_, err := NewBlockHeader(data)
	if err != nil {
		return errors.New("bytes do not correspond to a *wire.BlockHeader object")
	}

	*m = data
	return nil
}

func (m *BTCHeaderBytes) Size() int {
	bz, _ := m.Marshal()
	return len(bz)
}

func (m BTCHeaderBytes) ToBlockHeader() *wire.BlockHeader {
	header, err := NewBlockHeader(m)
	// There was a parsing error
	if err != nil {
		panic("BTCHeaderBytes cannot be converted to a block header object")
	}
	return header
}

func (m *BTCHeaderBytes) FromBlockHeader(header *wire.BlockHeader) {
	var buf bytes.Buffer
	err := header.Serialize(&buf)
	if err != nil {
		panic("*wire.BlockHeader cannot be serialized")
	}

	err = m.Unmarshal(buf.Bytes())
	if err != nil {
		panic("*wire.BlockHeader serialized bytes cannot be unmarshalled")
	}
}

func (m *BTCHeaderBytes) Eq(other *BTCHeaderBytes) bool {
	return m.Hash().Eq(other.Hash())
}

func (m *BTCHeaderBytes) Hash() *BTCHeaderHashBytes {
	blockHash := m.ToBlockHeader().BlockHash()
	hashBytes := NewBTCHeaderHashBytesFromChainhash(&blockHash)
	return &hashBytes
}

// NewBlockHeader creates a block header from bytes.
func NewBlockHeader(data []byte) (*wire.BlockHeader, error) {
	// Create an empty header
	header := &wire.BlockHeader{}

	// The Deserialize method expects an io.Reader instance
	reader := bytes.NewReader(data)
	// Decode the header bytes
	err := header.Deserialize(reader)
	return header, err
}
