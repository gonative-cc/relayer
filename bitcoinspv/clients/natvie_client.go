package clients

import (
	"context"
	"fmt"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

const (
	InsertHeader = "insert_header"
	VerifyTx     = "verify_tx"
	LcModule     = "light_client"
	LcPackage    = "0xec871f40adb592f9143351695a225c09b23a7dca6586a5ce1c6bf358de5ddc62"
)

type Client struct {
	c                   *sui.Client
	signerAccount       *signer.Signer
	lightClientObjectId string
}

func NewNativeClient(suiClient *sui.Client, signer *signer.Signer, lightClientObjectId string) Client {
	return Client{
		c:                   suiClient,
		signerAccount:       signer,
		lightClientObjectId: lightClientObjectId,
	}
}

func (c Client) InsertHeaders(ctx context.Context, blockHeader []*wire.BlockHeader) error {
	rawHeader, err := BlockHeaderToHex(*blockHeader[0])
	if err != nil {
		return fmt.Errorf("error serializing block header: %w", err)
	}

	arguments := []interface{}{
		c.lightClientObjectId, // The object ID of the light client
		rawHeader,             // The serialized block header as a hex string
	}

	req := models.MoveCallRequest{
		Signer:          c.signerAccount.Address,
		PackageObjectId: LcPackage,
		Module:          LcModule,
		Function:        InsertHeader,
		TypeArguments:   []interface{}{},
		Arguments:       arguments,
		GasBudget:       "100000000", // Adjust gas budget as needed
	}

	response, err := c.c.MoveCall(ctx, req)
	if err != nil {
		return fmt.Errorf("error calling insert_header function: %w", err)
	}

	_, err = c.c.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: response,
		PriKey:      c.signerAccount.PriKey,
		Options:     models.SuiTransactionBlockOptions{},
		RequestType: "WaitForLocalExecution",
	})
	if err != nil {
		return fmt.Errorf("error executing transaction: %w", err)
	}

	return nil
}

func (c Client) ContainsBTCBlock(blockHash chainhash.Hash) (bool, error) {
	// TODO: Implement logic to call the Sui smart contract method for checking block existence
	// Use nc.suiClient to interact with the Sui blockchain
	fmt.Println("ContainsBTCBlock called")
	return false, nil
}

func (c Client) GetHeaderChainTip() (Block, error) {
	// TODO: Implement logic to call the Sui smart contract method for fetching chain tip
	// Use nc.suiClient to interact with the Sui blockchain
	fmt.Println("BTCHeaderChainTip called")
	return Block{nil, 0}, nil
}

func (c Client) VerifySPV(spvProof *types.SPVProof) (int, error) {
	// TODO: Implement logic to call the Sui smart contract method for verifying SPV proofs
	// Use nc.suiClient to interact with the Sui blockchain
	fmt.Println("VerifySPV called")
	return 0, nil
}

func (c Client) Stop() {
	// TODO: Implement any necessary cleanup or shutdown logic
	fmt.Println("Stop called")
}
