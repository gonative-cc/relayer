package sui

// copied from our bitcoin light client repo
import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

const BTCHeaderSize = 80 // 80 bytes

// Utils for converting hex string to header
// The input must be 80 bytes hex string type
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
	buf := bytes.NewBuffer(make([]byte, 0, 80))
	err := header.Serialize(buf)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(buf.Bytes()), nil
}

// BlockHeaderToHex transforms Hash to a natural order hex endoed string
func BlockHashToHex(hash chainhash.Hash) string {
	return "0x" + hex.EncodeToString(hash.CloneBytes())
}
