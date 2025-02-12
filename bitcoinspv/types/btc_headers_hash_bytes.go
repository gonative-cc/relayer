package types

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type BTCHeaderHashBytes []byte

const BTCHeaderHashLen = 32

func NewBTCHeaderHashBytesFromChainhash(chHash *chainhash.Hash) BTCHeaderHashBytes {
	var headerHashBytes BTCHeaderHashBytes
	headerHashBytes.FromChainhash(chHash)
	return headerHashBytes
}

func (m BTCHeaderHashBytes) Marshal() ([]byte, error) {
	// Just return the bytes
	return m, nil
}

func (m *BTCHeaderHashBytes) Unmarshal(bz []byte) error {
	if len(bz) != BTCHeaderHashLen {
		return errors.New("invalid header hash length")
	}
	// Verify that the bytes can be transformed to a *chainhash.Hash object
	_, err := toChainhash(bz)
	if err != nil {
		return errors.New("bytes do not correspond to *chainhash.Hash object")
	}
	*m = bz
	return nil
}

func (m *BTCHeaderHashBytes) Size() int {
	bz, _ := m.Marshal()
	return len(bz)
}

func (m BTCHeaderHashBytes) ToChainhash() *chainhash.Hash {
	chHash, err := toChainhash(m)
	if err != nil {
		panic("BTCHeaderHashBytes cannot be converted to chainhash")
	}
	return chHash
}

func (m *BTCHeaderHashBytes) FromChainhash(hash *chainhash.Hash) {
	err := m.Unmarshal(hash[:])
	if err != nil {
		panic("*chainhash.Hash bytes cannot be unmarshalled")
	}
}

func (m *BTCHeaderHashBytes) String() string {
	return m.ToChainhash().String()
}

func (m *BTCHeaderHashBytes) Eq(hash *BTCHeaderHashBytes) bool {
	return m.String() == hash.String()
}

func toChainhash(data []byte) (*chainhash.Hash, error) {
	return chainhash.NewHash(data)
}
