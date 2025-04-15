package sui

import (
	"context"
	"testing"

	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// local variables used for testing
var (
	localRPC            = "http://127.0.0.1:9000"
	localMnemonic       = "hungry soft price stem lobster liar super protect script captain spring doctor"
	lightClientObjectID = "0xfdd31cc07afc6950aaee0ac85c45d244778b4b9c7f1ce91ccd39ada8c2731460"
	lcPackage           = "0x063d5eab5a5d09c22f1cf4e2dad1c91fd7172f72bea6a9d9a34939996fc84e2a"
)

func setupIntegrationTest(t *testing.T) (context.Context, clients.BitcoinSPV) {
	t.Helper()

	cl := sui.NewSuiClient(localRPC).(*sui.Client)
	s, err := signer.NewSignertWithMnemonic(localMnemonic)
	assert.Nil(t, err)
	client, err := NewSPVClient(
		cl,
		s,
		lightClientObjectID,
		lcPackage,
		zerolog.Logger{},
	)
	assert.Nil(t, err)

	return context.TODO(), client
}

func TestInsertHeader(t *testing.T) {
	t.Skip("Test to be run locally for debugging purposes only")
	ctx, client := setupIntegrationTest(t)

	rawHeaderHex := "000000307306011c31d1f14a422c50c70cbedb1233757505cb887d82d51ae3f27e23062d6be46c161e69696c1c83ba3a1ea52f071fcdada5a6bce28f5da591b969b42da139c5b167ffff7f2000000000"
	header, err := BlockHeaderFromHex(rawHeaderHex)
	assert.Nil(t, err)

	headers := []wire.BlockHeader{header}

	err = client.InsertHeaders(ctx, headers)
	assert.Nil(t, err)
}

func TestInsertHeaderAlreadyExistErr(t *testing.T) {
	t.Skip("Test to be run locally for debugging purposes only")
	ctx, client := setupIntegrationTest(t)

	rawHeaderHex := "00000030759e91f85448e42780695a7c71a6e4f4e845ecd895b19fafaeb6f5e3c030e62233287429255f254a463d90b998ba5523634da7c67ef873268e1db40d1526d5583d5b6167ffff7f2000000000"
	header, err := BlockHeaderFromHex(rawHeaderHex)
	assert.Nil(t, err)

	headers := []wire.BlockHeader{header}

	err = client.InsertHeaders(ctx, headers)
	assert.NotNil(t, err)
}

func TestContainsBlock(t *testing.T) {
	t.Skip("Test to be run locally for debugging purposes only")
	ctx, client := setupIntegrationTest(t)

	rawHeaderHex := "00000030759e91f85448e42780695a7c71a6e4f4e845ecd895b19fafaeb6f5e3c030e62233287429255f254a463d90b998ba5523634da7c67ef873268e1db40d1526d5583d5b6167ffff7f2000000000"
	header, err := BlockHeaderFromHex(rawHeaderHex)
	assert.Nil(t, err)

	exist, err := client.ContainsBlock(ctx, header.BlockHash())
	assert.Nil(t, err)
	assert.True(t, exist, "Block should exist")

	nonExistentHash, _ := chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000001")

	exist, err = client.ContainsBlock(ctx, *nonExistentHash)
	assert.Nil(t, err)
	assert.False(t, exist, "Non-existent block should not exist")
}

func TestGetHeaderChainTip(t *testing.T) {
	t.Skip("Test to be run locally for debugging purposes only")
	ctx, client := setupIntegrationTest(t)

	blockInfo, err := client.GetLatestBlockInfo(ctx)
	assert.Nil(t, err)
	assert.NotZero(t, blockInfo.Height)

	exist, err := client.ContainsBlock(ctx, *blockInfo.Hash)
	assert.Nil(t, err)
	assert.True(t, exist, "Chain tip block should exist")
}
