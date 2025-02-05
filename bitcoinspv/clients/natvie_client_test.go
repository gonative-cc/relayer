package clients

import (
	"context"
	"testing"

	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/stretchr/testify/assert"
)

func TestNatvieClient_InsertHeader(t *testing.T) {
	t.Skip("Test to be run locally for debugging purposes only")

	localRPC := "127.0.0.1:2000"
	localMnemonic := "local mnemonic here"
	lightClientObjectId := "0x..."
	gasObjectId := "0x..."

	cl := sui.NewSuiClient(localRPC).(*sui.Client)
	s, err := signer.NewSignertWithMnemonic(localMnemonic)
	assert.Nil(t, err)

	client := NewNativeClient(
		cl,
		s,
		lightClientObjectId,
		gasObjectId,
	)
	assert.Nil(t, err)

	rawHeaderHex := "00801e31c24ae25304cbac7c3d3b076e241abb20ff2da1d3ddfc00000000000000000000530e6745eca48e937428b0f15669efdce807a071703ed5a4df0e85a3f6cc0f601c35cf665b25031780f1e351"

	header, err := BlockHeaderFromHex(rawHeaderHex)
	assert.Nil(t, err)

	err = client.InsertHeaders(
		context.Background(),
		&header,
	)

	assert.Nil(t, err)
}
