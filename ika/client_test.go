package ika

import (
	"context"
	"os"
	"testing"

	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestClient_ApproveAndSign(t *testing.T) {
	t.Skip("Test to be run locally for debugging purposes only")
	err := godotenv.Load("./../.env.test")
	assert.Nil(t, err)

	localRPC := os.Getenv("LOCAL_RPC")
	localMnemonic := os.Getenv("LOCAL_MNEMONIC")
	lcPackage := os.Getenv("LC_PACKAGE")
	lcModule := os.Getenv("LC_MODULE")
	gasAccount := os.Getenv("IKA_GAS_ACC")
	gasBudget := os.Getenv("IKA_GAS_BUDGET")
	lcFunction := "test"

	cl := sui.NewSuiClient(localRPC).(*sui.Client)
	s, err := signer.NewSignertWithMnemonic(localMnemonic)
	assert.Nil(t, err)

	client, err := NewClient(
		cl,
		s,
		SuiCtrCall{
			Package:  lcPackage,
			Module:   lcModule,
			Function: lcFunction,
		},
		gasAccount,
		gasBudget,
	)
	assert.Nil(t, err)

	signature, err := client.ApproveAndSign(
		context.Background(),
		"0x62e79d33bb331d8f93973252b4d2eda5a491d9b87044808530b25c60fc98b276",
		"0x62e79d33bb331d8f93973252b4d2eda5a491d9b87044808530b25c60fc98b276",
		[][]byte{{1, 2, 3}}, // Mock messages
	)

	assert.Nil(t, err)
	assert.Equal(t, len(signature), 1)
	assert.Equal(t, signature[0], []byte("mock_signature"))
}
