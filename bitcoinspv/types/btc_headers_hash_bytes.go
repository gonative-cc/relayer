package types

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

const BTCHeaderHashLen = 32

type BTCHeaderHashBytes []byte

func (b BTCHeaderHashBytes) Marshal() ([]byte, error) {
	return b, nil
}

func (b *BTCHeaderHashBytes) Unmarshal(data []byte) error {
	if len(data) != BTCHeaderHashLen {
		return errors.New("header hash must be exactly 32 bytes")
	}

	if _, err := toChainhash(data); err != nil {
		return errors.New("failed to convert bytes to chainhash.Hash format")
	}

	*b = data
	return nil
}

func (b *BTCHeaderHashBytes) Size() int {
	data, _ := b.Marshal()
	return len(data)
}

func NewBTCHeaderHashBytesFromChainhash(hash *chainhash.Hash) BTCHeaderHashBytes {
	result := BTCHeaderHashBytes{}
	result.FromChainhash(hash)
	return result
}

func (b BTCHeaderHashBytes) ToChainhash() *chainhash.Hash {
	result, err := toChainhash(b)
	if err != nil {
		panic("failed to convert BTCHeaderHashBytes to chainhash format")
	}
	return result
}

func (b *BTCHeaderHashBytes) FromChainhash(hash *chainhash.Hash) {
	if err := b.Unmarshal(hash[:]); err != nil {
		panic("failed to convert chainhash.Hash bytes to BTCHeaderHashBytes")
	}
}

func (b *BTCHeaderHashBytes) String() string {
	return b.ToChainhash().String()
}

func (b *BTCHeaderHashBytes) Eq(other *BTCHeaderHashBytes) bool {
	return b.String() == other.String()
}

func toChainhash(data []byte) (*chainhash.Hash, error) {
	return chainhash.NewHash(data)
}
