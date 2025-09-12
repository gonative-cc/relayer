package sui

// copied from our bitcoin light client repo
import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// BTCHeaderSize is the size in bytes of a Bitcoin block header.
const BTCHeaderSize = 80

// BlockHeader is block header
type BlockHeader struct {
	Internal []uint8
}

// LightBlock is light block
//
//nolint:govet
type LightBlock struct {
	Height    uint64
	ChainWork [32]uint8
	Header    *BlockHeader
}

// BlockHeaderFromHex converts a hexadecimal string representation of a Bitcoin
// block header into a wire.BlockHeader struct.
// The input must be 80 bytes hex string type.
func BlockHeaderFromHex(hexStr string) (wire.BlockHeader, error) {
	var header wire.BlockHeader
	if len(hexStr) != BTCHeaderSize*2 {
		return header, errors.New("invalid header size, must be 80 bytes")
	}

	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return header, err
	}

	reader := bytes.NewReader(data)
	err = header.Deserialize(reader)
	return header, err
}

// BlockHeaderToHex transforms header to hex encoded string
func BlockHeaderToHex(header wire.BlockHeader) (string, error) {
	buf := bytes.NewBuffer(make([]byte, 0, BTCHeaderSize))
	err := header.Serialize(buf)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(buf.Bytes()), nil
}

// BlockHashToHex transforms Hash to a natural order hex endoed string
func BlockHashToHex(hash chainhash.Hash) string {
	return "0x" + hex.EncodeToString(hash.CloneBytes())
}

// BlockHash returns block hash
func (lb LightBlock) BlockHash() (chainhash.Hash, error) {
	r := bytes.NewReader(lb.Header.Internal)
	var header wire.BlockHeader
	err := header.Deserialize(r)
	if err != nil {
		return header.BlockHash(), err
	}
	return header.BlockHash(), nil
}

func blockHeaderToBytes(header wire.BlockHeader) ([]byte, error) {
	var w bytes.Buffer
	if err := header.Serialize(&w); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
