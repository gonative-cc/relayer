package remote2ika

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/pattonkan/sui-go/suiclient"
	"github.com/pattonkan/sui-go/suisigner"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	t.Skip("Test to be run locally for debugging purposes only")
	err := godotenv.Load("./../.env.test")
	assert.Nil(t, err)

	localRPC := os.Getenv("LOCAL_RPC")
	localMnemonic := os.Getenv("LOCAL_MNEMONIC")
	dwalletPackage := os.Getenv("LC_PACKAGE")
	dWalletModule := os.Getenv("LC_MODULE")
	gasAccount := os.Getenv("IKA_GAS_ACC")
	gasBudget := os.Getenv("IKA_GAS_BUDGET")
	spvLCFun := "test"

	cl := suiclient.NewClient(localRPC)
	s, err := suisigner.NewSignerWithMnemonic(localMnemonic, suisigner.KeySchemeFlagDefault)
	assert.Nil(t, err)

	client, err := NewClient(
		cl,
		s,
		SuiCtrCall{
			Package:  dwalletPackage,
			Module:   dWalletModule,
			Function: spvLCFun,
		},
		SuiCtrCall{
			Package:  dwalletPackage,
			Module:   dWalletModule,
			Function: spvLCFun,
		},
		gasAccount,
		gasBudget,
	)
	assert.Nil(t, err)

	txDigest, err := client.SignReq(
		context.Background(),
		"0x62e79d33bb331d8f93973252b4d2eda5a491d9b87044808530b25c60fc98b276",
		"0x62e79d33bb331d8f93973252b4d2eda5a491d9b87044808530b25c60fc98b276",
		[][]byte{{1, 2, 3}}, // Mock messages
	)

	assert.Nil(t, err)
	assert.Equal(t, txDigest, "txDigest")
}
