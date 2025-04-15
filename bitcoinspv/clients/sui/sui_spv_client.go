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
	"github.com/rs/zerolog"
)

const (
	insertHeadersFunc = "insert_headers"
	containsBlockFunc = "exist"
	getChainTipFunc   = "latest_block_hash"
	verifySPVFunc     = "verify_tx"
	lcModule          = "light_client"
	defaultGasBudget  = "10000000000"
)

// SPVClient implements the BitcoinSPV interface, interacting with a
// Bitcoin SPV light client deployed as a smart contract on Sui.
type SPVClient struct {
	logger      zerolog.Logger
	suiClient   *sui.Client
	signer      *signer.Signer
	lcObjectID  string
	lcPackageID string
}

// NewSPVClient creates a new SPVClient instance.
func NewSPVClient(
	suiClient *sui.Client,
	signer *signer.Signer,
	lightClientObjectID string,
	lightClientPackageID string,
	parentLogger zerolog.Logger,
) (clients.BitcoinSPV, error) {
	if suiClient == nil {
		return nil, ErrSuiClientNil
	}
	if signer == nil {
		return nil, ErrSignerNill
	}
	if lightClientObjectID == "" || lightClientPackageID == "" {
		return nil, ErrEmptyObjectID
	}

	return &SPVClient{
		suiClient:   suiClient,
		signer:      signer,
		lcObjectID:  lightClientObjectID,
		lcPackageID: lightClientPackageID,
		logger:      configureClientLogger(parentLogger),
	}, nil
}

func configureClientLogger(parentLogger zerolog.Logger) zerolog.Logger {
	return parentLogger.With().Str("module", "spv_client").Logger()
}

// InsertHeaders adds new Bitcoin block headers to the light client's chain.
func (c *SPVClient) InsertHeaders(ctx context.Context, blockHeaders []wire.BlockHeader) error {
	if len(blockHeaders) == 0 {
		return ErrNoBlockHeaders
	}

	rawHeaders := make([]string, 0, len(blockHeaders))
	for _, header := range blockHeaders {
		rawHeader, err := BlockHeaderToHex(header)
		if err != nil {
			return fmt.Errorf("error serializing block header: %w", err)
		}
		rawHeaders = append(rawHeaders, rawHeader)
	}

	arguments := []interface{}{
		c.lcObjectID,
		rawHeaders,
	}

	c.logger.Debug().Msgf("Calling insert headers with the following arguemts: %v", arguments...)

	_, err := c.moveCall(ctx, insertHeadersFunc, arguments)
	return err
}

// ContainsBlock checks if the light client's chain includes a block with the given hash.
func (c *SPVClient) ContainsBlock(ctx context.Context, blockHash chainhash.Hash) (bool, error) {
	arguments := []interface{}{
		c.lcObjectID,
		BlockHashToHex(blockHash),
	}

	response, err := c.moveCall(ctx, containsBlockFunc, arguments)
	if err != nil {
		return false, err
	}

	eventData, err := c.extractFirstEvent(ctx, response.Effects.TransactionDigest)
	if err != nil {
		return false, err
	}
	exist, ok := eventData["exist"].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected event data format: 'exist' field not found or not a boolean")
	}

	return exist, nil
}

// GetLatestBlockInfo returns the block hash and height of the best block header.
func (c *SPVClient) GetLatestBlockInfo(ctx context.Context) (*clients.BlockInfo, error) {
	arguments := []interface{}{
		c.lcObjectID,
	}

	response, err := c.moveCall(ctx, getChainTipFunc, arguments)
	if err != nil {
		return nil, err
	}

	eventData, err := c.extractFirstEvent(ctx, response.Effects.TransactionDigest)
	if err != nil {
		return nil, err
	}

	lightBlockHashData, ok := eventData["light_block_hash"].([]interface{})
	if !ok {
		return nil, ErrLightBlockHashNotFound
	}

	heightData, ok := eventData["height"]
	if !ok {
		return nil, ErrHeightNotFound
	}

	heightStr, ok := heightData.(string)
	if !ok {
		return nil, fmt.Errorf("%w: got %T", ErrHeightInvalidType, heightData)
	}

	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrHeightInvalidValue, err.Error())
	}

	hashBytes, err := extractHashBytes(lightBlockHashData)
	if err != nil {
		return nil, err
	}

	blockHash, err := chainhash.NewHash(hashBytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBlockHashInvalid, err)
	}

	return &clients.BlockInfo{
		Hash:   blockHash,
		Height: height,
	}, nil
}

func extractHashBytes(lightBlockHashData []interface{}) ([]byte, error) {
	hashBytes := make([]byte, len(lightBlockHashData))
	for i, v := range lightBlockHashData {
		byteVal, ok := v.(float64) // JSON numbers come as float64.
		if !ok {
			return nil,
				fmt.Errorf("%w: got %T at index %d", ErrBlockHashInvalidType, v, i)
		}
		if byteVal < 0 || byteVal > 255 {
			return nil, fmt.Errorf("%w: %f at index %d", ErrBlockHashInvalidByte, byteVal, i)
		}
		hashBytes[i] = byte(byteVal)
	}
	return hashBytes, nil
}

// VerifySPV verifies an SPV proof against the light client's stored headers.
// TODO: finish implementation
func (c *SPVClient) VerifySPV(ctx context.Context, spvProof *types.SPVProof) (int, error) {
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
	eventData, err := c.extractFirstEvent(ctx, response.Effects.TransactionDigest)
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

// Stop performs any necessary cleanup and shutdown operations.
func (c SPVClient) Stop() {
	// TODO: Implement any necessary cleanup or shutdown logic
	fmt.Println("Stop called")
}

// moveCall is a helper function to construct and execute a Move call on the Sui blockchain.
func (c *SPVClient) moveCall(
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
			fmt.Errorf("sui transaction submission for '%s' failed: %w", function, err)
	}

	// The error returned by SignAndExecuteTransactionBlock only indicates
	// whether the transaction was successfully submitted to the network.
	// It does NOT guarantee that the transaction succeeded  during execution.
	// Thats why we MUST inspect the `Effects.Status` field.
	// It will tell us about execution errors like: Abort, OutOfGas etc.
	if signedResp.Effects.Status.Status != "success" {
		return signedResp, fmt.Errorf("%w: function '%s' status: %s, error: %s",
			ErrSuiTransactionFailed, function, signedResp.Effects.Status.Status, signedResp.Effects.Status.Error)
	}

	return signedResp, nil
}

// extractFirstEvent is a helper function to extract data from the first event in a transaction.
//
//	It returns error if there is no events.
func (c *SPVClient) extractFirstEvent(ctx context.Context, txDigest string) (map[string]interface{}, error) {
	events, err := c.suiClient.SuiGetEvents(ctx, models.SuiGetEventsRequest{
		Digest: txDigest,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEventDataFormat, err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoEventsFound, txDigest)
	}

	return events[0].ParsedJson, nil
}
