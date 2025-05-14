package suigoclient

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/wire"
	"github.com/pattonkan/sui-go/suiclient"
)

const btcHeaderSize = 80

type bcsEncode []byte

func getBCSResult(res *suiclient.DevInspectTransactionBlockResponse) ([]bcsEncode, error) {
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
	return bcsEncode, nil
}

// BlockHeaderToHex transforms header to hex encoded string
func toBytes(header wire.BlockHeader) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, btcHeaderSize))
	err := header.Serialize(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BlockHeaderToHex transforms header to hex encoded string
func BlockHeaderToHex(header wire.BlockHeader) (string, error) {
	buf := bytes.NewBuffer(make([]byte, 0, btcHeaderSize))
	err := header.Serialize(buf)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(buf.Bytes()), nil
}
