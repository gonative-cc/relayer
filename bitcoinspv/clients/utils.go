package clients

//copied from our bitcoin light client repo
import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

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

func serializeBlockHeader(header *wire.BlockHeader) ([]byte, error) {
	// Use JSON serialization as an example
	// You can replace this with any other serialization method (e.g., gob, protobuf, etc.)
	rawHeader, err := json.Marshal(header)
	if err != nil {
		return nil, fmt.Errorf("error marshaling block header: %w", err)
	}
	return rawHeader, nil
}
