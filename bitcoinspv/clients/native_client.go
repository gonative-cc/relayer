package clients

import (
	"context"
	"encoding/hex"
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
	LcModule     = "bitcoin_spv"
	LcPackage    = "0x<your_package_object_id>"
)

type Client struct {
	c                   *sui.Client
	signerAccount       *signer.Signer
	lightClientObjectId string
	gasObjectId         string
}

func NewNativeClient(suiClient *sui.Client, signer *signer.Signer, lightClientObjectId string, gasObjectId string) Client {
	return Client{
		c:                   suiClient,
		signerAccount:       signer,
		lightClientObjectId: lightClientObjectId,
		gasObjectId:         gasObjectId,
	}
}

func (c *Client) InsertHeaders(ctx context.Context, blockHeader *wire.BlockHeader) error {
	rawHeader, err := serializeBlockHeader(blockHeader)
	if err != nil {
		return fmt.Errorf("error serializing block header: %w", err)
	}
	rawHeaderHex := hex.EncodeToString(rawHeader)

	arguments := []interface{}{
		c.lightClientObjectId, // The object ID of the light client
		rawHeaderHex,          // The serialized block header as a hex string
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

func (c *Client) ContainsBTCBlock(blockHash *chainhash.Hash) (bool, error) {
	// TODO: Implement logic to call the Sui smart contract method for checking block existence
	// Use nc.suiClient to interact with the Sui blockchain
	fmt.Println("ContainsBTCBlock called")
	return false, nil
}

func (c *Client) BTCHeaderChainTip() (int64, *chainhash.Hash, error) {
	// TODO: Implement logic to call the Sui smart contract method for fetching chain tip
	// Use nc.suiClient to interact with the Sui blockchain
	fmt.Println("BTCHeaderChainTip called")
	return 0, nil, nil
}

func (c *Client) VerifySPV(spvProof types.SPVProof) (int, error) {
	// TODO: Implement logic to call the Sui smart contract method for verifying SPV proofs
	// Use nc.suiClient to interact with the Sui blockchain
	fmt.Println("VerifySPV called")
	return 0, nil
}

func (c *Client) Stop() error {
	// TODO: Implement any necessary cleanup or shutdown logic
	fmt.Println("Stop called")
	return nil
}
