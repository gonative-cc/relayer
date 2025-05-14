package sui

// copied from our bitcoin light client repo
import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/pattonkan/sui-go/suiclient"
)

// BTCHeaderSize is the size in bytes of a Bitcoin block header.
const BTCHeaderSize = 80

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

type bcsEncode []byte

func getBCSResult(res *suiclient.DevInspectTransactionBlockResponse) []bcsEncode {
	bcsEncode := make([]bcsEncode, len(res.Results[0].ReturnValues))

	for i, item := range res.Results[0].ReturnValues {
		var b []byte
		// TODO: Breakdown to simple term
		c := item.([]interface{})[0].([]interface{})
		b = make([]byte, len(c))

		for i, v := range c {
			b[i] = byte(v.(float64))
		}
		bcsEncode[i] = b
	}
	return bcsEncode
}

// BlockHeader is block header
type BlockHeader struct {
	Internal []uint8
}

// LightBlock is light block
// TODO: fix lint
//
//nolint:govet
type LightBlock struct {
	Height    uint64
	ChainWork [32]uint8
	Header    *BlockHeader
}

// BlockHash return block hash
func (lb LightBlock) BlockHash() (chainhash.Hash, error) {
	r := bytes.NewReader(lb.Header.Internal)
	var header wire.BlockHeader
	err := header.Deserialize(r)
	if err != nil {
		return header.BlockHash(), err
	}
	return header.BlockHash(), nil
}
