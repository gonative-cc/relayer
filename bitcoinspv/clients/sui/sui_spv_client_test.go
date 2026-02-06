package sui

import (
	"context"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/fardream/go-bcs/bcs"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/pattonkan/sui-go/sui"
	"github.com/pattonkan/sui-go/suiclient"
	"github.com/pattonkan/sui-go/suisigner"
	"github.com/pattonkan/sui-go/suisigner/suicrypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// local variables used for testing
var (
	localRPC            = "http://127.0.0.1:9000"
	localMnemonic       = "hungry soft price stem lobster liar super protect script captain spring doctor"
	lightClientObjectID = "0xfdd31cc07afc6950aaee0ac85c45d244778b4b9c7f1ce91ccd39ada8c2731460"
	lcPkgID             = "0x063d5eab5a5d09c22f1cf4e2dad1c91fd7172f72bea6a9d9a34939996fc84e2a"
	btcLibPkg           = "0xf7d3be2ce8504a3fb5999ef46d6725e1024b34cf6cfecdb3d2b5645b0a98c55d"
)

func setupIntegrationTest(t *testing.T) (context.Context, clients.BitcoinSPV) {
	t.Helper()

	cl := suiclient.NewClient(localRPC)
	s, err := suisigner.NewSignerWithMnemonic(localMnemonic, suicrypto.KeySchemeFlagDefault)
	assert.Nil(t, err)
	client, err := New(
		cl,
		s,
		lightClientObjectID,
		lcPkgID,
		btcLibPkg,
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

func TestParseHeight(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		height := uint64(12345)
		data, _ := bcs.Marshal(height)
		val := suiclient.ReturnValueType{
			Data:    data,
			TypeTag: &sui.TypeTag{U64: &sui.EmptyEnum{}},
		}

		result, err := parseHeight(val)
		assert.NoError(t, err)
		assert.Equal(t, height, result)
	})

	t.Run("wrong type tag", func(t *testing.T) {
		val := suiclient.ReturnValueType{
			TypeTag: &sui.TypeTag{U8: &sui.EmptyEnum{}},
		}
		_, err := parseHeight(val)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected return type")
	})

	t.Run("unmarshal failure", func(t *testing.T) {
		val := suiclient.ReturnValueType{
			Data:    []byte{1, 2}, // too short for u64
			TypeTag: &sui.TypeTag{U64: &sui.EmptyEnum{}},
		}
		_, err := parseHeight(val)
		assert.Error(t, err)
	})
}

func TestParseHash(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		hashStr := "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"
		hash, _ := chainhash.NewHashFromStr(hashStr)
		data, _ := bcs.Marshal(hash[:])
		val := suiclient.ReturnValueType{
			Data: data,
			TypeTag: &sui.TypeTag{Vector: &sui.TypeTag{
				U8: &sui.EmptyEnum{},
			}},
		}

		result, err := parseHash(val)
		assert.NoError(t, err)
		assert.Equal(t, hash, result)
	})

	t.Run("wrong type tag", func(t *testing.T) {
		val := suiclient.ReturnValueType{
			TypeTag: &sui.TypeTag{U64: &sui.EmptyEnum{}},
		}
		_, err := parseHash(val)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected return type")
	})

	t.Run("invalid hash length", func(t *testing.T) {
		data, _ := bcs.Marshal([]byte{1, 2, 3})
		val := suiclient.ReturnValueType{
			Data: data,
			TypeTag: &sui.TypeTag{Vector: &sui.TypeTag{
				U8: &sui.EmptyEnum{},
			}},
		}
		_, err := parseHash(val)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create chainhash")
	})
}
