package clients

import (
	"context"
	"testing"

	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
)

func TestNatvieClient_InsertHeader(t *testing.T) {
	t.Skip("Test to be run locally for debugging purposes only")
	localRPC := "https://fullnode.devnet.sui.io:443"
	localMnemonic := "elegant aware place green laptop secret misery mass crystal cash armor opera"
	lightClientObjectID := "0x11c3a8e5848a516b50fbccf4d7504aa4e9fc1fe7c29493b7c951c417349da8d1"

	cl := sui.NewSuiClient(localRPC).(*sui.Client)
	s, err := signer.NewSignertWithMnemonic(localMnemonic)
	assert.Nil(t, err)

	client := NewNativeClient(
		cl,
		s,
		lightClientObjectID,
	)
	assert.Nil(t, err)

	rawHeaderHex := "0000002038e3369b8a033550a8443c3e0f51c0110c7d527d98437da1ca24557147b38c4a0530f86490bb70e5a8233a2f9b7c3a525bd7267498d4b4efa2bf2ff00cd988f21d5b6167ffff7f2000000000"

	header, err := BlockHeaderFromHex(rawHeaderHex)
	assert.Nil(t, err)

	var headers []*wire.BlockHeader = []*wire.BlockHeader{&header}

	err = client.InsertHeaders(
		context.Background(),
		headers,
	)

	assert.Nil(t, err)
}
