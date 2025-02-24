package sui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/rs/zerolog/log"
)

const (
	insertHeadersFunc = "insert_headers"
	containsBlockFunc = "exist"
	getChainTipFunc   = "latest_block_hash"
	verifySPVFunc     = "verify_tx"
	lcModule          = "light_client"
	defaultGasBudget  = "100000000"
)

// BitcoinSPVClient implements the BitcoinSPVClient interface, interacting with a
// Bitcoin SPV light client deployed as a smart contract on Sui.
type BitcoinSPVClient struct {
	suiClient   *sui.Client
	signer      *signer.Signer
	lcObjectID  string
	lcPackageID string
}

// NewBitcoinSPVClient creates a new BitcoinSPVClient instance.
func NewBitcoinSPVClient(
	suiClient *sui.Client,
	signer *signer.Signer,
	lightClientObjectID string,
	lightClientPackageID string,
) (*BitcoinSPVClient, error) {
	if suiClient == nil {
		return nil, ErrSuiClientNil
	}
	if signer == nil {
		return nil, ErrSignerNill
	}
	if lightClientObjectID == "" || lightClientPackageID == "" {
		return nil, ErrEmptyObjectID
	}

	return &BitcoinSPVClient{
		suiClient:   suiClient,
		signer:      signer,
		lcObjectID:  lightClientObjectID,
		lcPackageID: lightClientPackageID,
	}, nil
}

// InsertHeaders adds new Bitcoin block headers to the light client's chain.
func (c BitcoinSPVClient) InsertHeaders(ctx context.Context, blockHeaders []*wire.BlockHeader) error {
	if len(blockHeaders) == 0 {
		return fmt.Errorf("no block headers provided")
	}

	rawHeaders := make([]string, 0, len(blockHeaders))
	for _, header := range blockHeaders {
		rawHeader, err := BlockHeaderToHex(*header)
		if err != nil {
			return fmt.Errorf("error serializing block header: %w", err)
		}
		rawHeaders = append(rawHeaders, rawHeader)
	}

	arguments := []interface{}{
		c.lcObjectID,
		rawHeaders,
	}

	_, err := c.moveCall(ctx, insertHeadersFunc, arguments)
	return err

}

// ContainsBTCBlock checks if the light client's chain includes a block with the given hash.
func (c *BitcoinSPVClient) ContainsBTCBlock(ctx context.Context, blockHash chainhash.Hash) (bool, error) {
	arguments := []interface{}{
		c.lcObjectID,
		BlockHashToHex(blockHash),
	}
	response, err := c.moveCall(ctx, containsBlockFunc, arguments)
	if err != nil {
		return false, err
	}

	eventData, err := c.extractEventData(ctx, response.Effects.TransactionDigest)
	if err != nil {
		return false, err
	}
	exist, ok := eventData["exist"].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected event data format: 'exist' field not found or not a boolean")
	}

	return exist, nil
}

// GetHeaderChainTip returns the block hash and height of the best block header.
func (c *BitcoinSPVClient) GetHeaderChainTip(ctx context.Context) (*clients.BlockInfo, error) {
	arguments := []interface{}{
		c.lcObjectID,
	}
	response, err := c.moveCall(ctx, getChainTipFunc, arguments)
	if err != nil {
		return nil, err
	}

	eventData, err := c.extractEventData(ctx, response.Effects.TransactionDigest)
	if err != nil {
		return nil, err
	}

	lightBlockHashData, ok := eventData["light_block_hash"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected event data format: 'light_block_hash' field not found or not a slice")
	}

	heightData, ok := eventData["height"]
	if !ok {
		return nil, fmt.Errorf("unexpected event data format: 'height' field not found")
	}

	heightStr, ok := heightData.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected event data format: 'height' expected type of string got %T", heightData)
	}

	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid height value '%s': %w", heightStr, err)
	}

	hashBytes := make([]byte, len(lightBlockHashData))
	for i, v := range lightBlockHashData {
		byteVal, ok := v.(float64) // JSON numbers come as float64.
		if !ok {
			return nil,
				fmt.Errorf("unexpected type in 'light_block_hash' array: expected float64, got %T at index %d", v, i)
		}
		if byteVal < 0 || byteVal > 255 {
			return nil, fmt.Errorf("invalid byte value in 'light_block_hash' array: %f at index %d", byteVal, i)
		}
		hashBytes[i] = byte(byteVal)
	}

	blockHash, err := chainhash.NewHash(hashBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid block hash bytes: %w", err)
	}

	return &clients.BlockInfo{
		Hash:   blockHash,
		Height: height,
	}, nil
}

// VerifySPV verifies an SPV proof against the light client's stored headers.
// TODO: finish implementation
func (c *BitcoinSPVClient) VerifySPV(ctx context.Context, spvProof *types.SPVProof) (int, error) {
	arguments := []interface{}{
		c.lcObjectID,
		spvProof.TxID,
		spvProof.MerklePath,
		spvProof.TxIndex,
	}

	response, err := c.moveCall(ctx, verifySPVFunc, arguments)
	if err != nil {
		return -1, err
	}
	eventData, err := c.extractEventData(ctx, response.Effects.TransactionDigest)
	if err != nil {
		return -1, err
	}

	// TODO: Define constants for the different verification states.
	result, ok := eventData["result"].(bool)
	if !ok {
		return -1, fmt.Errorf("unexpected event data format: 'result' field not found or not a boolean")
	}

	if result {
		return 2, nil
	}
	return 1, nil
}

func (c BitcoinSPVClient) Stop() {
	// TODO: Implement any necessary cleanup or shutdown logic
	fmt.Println("Stop called")
}

// moveCall is a helper function to construct and execute a Move call on the Sui blockchain.
func (c *BitcoinSPVClient) moveCall(
	ctx context.Context,
	function string,
	arguments []interface{},
) (models.SuiTransactionBlockResponse, error) {
	req := models.MoveCallRequest{
		Signer:          c.signer.Address,
		PackageObjectId: c.lcPackageID,
		Module:          lcModule,
		Function:        function,
		TypeArguments:   []interface{}{},
		Arguments:       arguments,
		GasBudget:       defaultGasBudget,
	}

	log.Debug().Msgf("SuiSPVClient: calling insert headers with the following arguemts: %v", arguments...)

	resp, err := c.suiClient.MoveCall(ctx, req)
	if err != nil {
		return models.SuiTransactionBlockResponse{}, fmt.Errorf("sui move call to '%s' failed: %w", function, err)
	}

	signedResp, err := c.suiClient.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: resp,
		PriKey:      c.signer.PriKey,
		Options: models.SuiTransactionBlockOptions{
			ShowInput:    true,
			ShowRawInput: true,
			ShowEffects:  true,
		},
		RequestType: "WaitForLocalExecution",
	})
	if err != nil {
		return models.SuiTransactionBlockResponse{},
			fmt.Errorf("sui transaction execution for '%s' failed: %w", function, err)
	}

	return signedResp, nil
}

// extractEventData is a helper function to extract data from the first event in a transaction.
//
//	It returns error if there is no events.
func (c *BitcoinSPVClient) extractEventData(ctx context.Context, txDigest string) (map[string]interface{}, error) {
	events, err := c.suiClient.SuiGetEvents(ctx, models.SuiGetEventsRequest{
		Digest: txDigest,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Sui events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for transaction digest: %s", txDigest)
	}

	// TODO: use raw JSON
	return events[0].ParsedJson, nil
}
