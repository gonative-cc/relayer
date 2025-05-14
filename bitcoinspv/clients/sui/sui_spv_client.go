package sui

import (
	"context"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/fardream/go-bcs/bcs"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/pattonkan/sui-go/sui"
	"github.com/pattonkan/sui-go/sui/suiptb"
	"github.com/pattonkan/sui-go/suiclient"
	"github.com/pattonkan/sui-go/suisigner"
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
// Bitcoin SPV light client deployed as a smart contract on Sui.logger
type LCClient struct {
	*suiclient.ClientImpl
	*suisigner.Signer
	PackageID *sui.PackageId
	LcObjArg  suiptb.CallArg
	logger    zerolog.Logger
}

var _ clients.BitcoinSPV = &LCClient{}

// NewSPVClient creates a new SPVClient instance.
func New(
	suiClient *suiclient.ClientImpl,
	signer *suisigner.Signer,
	lightClientObjectIDHex string,
	lightClientPackageIDHex string,
	parentLogger zerolog.Logger,
) (clients.BitcoinSPV, error) {
	if suiClient == nil {
		return nil, ErrSuiClientNil
	}
	if signer == nil {
		return nil, ErrSignerNill
	}

	lcPackageID, err := sui.PackageIdFromHex(lightClientPackageIDHex)
	if err != nil {
		return nil, err
	}

	lcObjectID, err := sui.ObjectIdFromHex(lightClientObjectIDHex)
	if err != nil {
		return nil, err
	}

	// tmp ctx
	// TODO: handle this in other PR
	ctx := context.TODO()
	lcObjectResp, err := suiClient.GetObject(ctx, &suiclient.GetObjectRequest{
		ObjectId: lcObjectID,
		Options: &suiclient.SuiObjectDataOptions{
			ShowOwner: true,
		},
	})

	lcObjArg := suiptb.CallArg{
		Object: &suiptb.ObjectArg{
			SharedObject: &suiptb.SharedObjectArg{
				Id:                   lcObjectResp.Data.ObjectId,
				InitialSharedVersion: *lcObjectResp.Data.Owner.Shared.InitialSharedVersion,
				Mutable:              false,
			},
		},
	}
	if err != nil {
		return nil, err
	}

	return &LCClient{
		ClientImpl: suiClient,
		Signer:     signer,
		logger:     configureClientLogger(parentLogger),
		PackageID:  lcPackageID,
		LcObjArg:   lcObjArg,
	}, nil
}

func configureClientLogger(parentLogger zerolog.Logger) zerolog.Logger {
	return parentLogger.With().Str("module", "spv_client").Logger()
}

// InsertHeaders adds new Bitcoin block headers to the light client's chain.
func (c *LCClient) InsertHeaders(ctx context.Context, blockHeaders []wire.BlockHeader) error {
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

	arguments := []any{
		c.LcObjArg.Object.SharedObject.Id,
		rawHeaders,
	}

	c.logger.Debug().Msgf("Calling insert headers with the following arguemts: %v", arguments...)

	_, err := c.moveCall(ctx, "insert_headers", arguments)
	return err
}

// ContainsBlock checks if the light client's chain includes a block with the given hash.
func (c *LCClient) ContainsBlock(ctx context.Context, blockHash chainhash.Hash) (bool, error) {
	ptb := suiptb.NewTransactionDataTransactionBuilder()

	b, err := bcs.Marshal(blockHash[:])
	if err != nil {
		return false, err
	}

	err = ptb.MoveCall(
		c.PackageID,
		"light_client",
		"exist",
		[]sui.TypeTag{},
		[]suiptb.CallArg{c.LcObjArg, {Pure: &b}},
	)
	if err != nil {
		return false, err
	}

	resp, err := c.devInspectTransactionBlock(ctx, ptb)

	if resp.Error != "" {
		return false, errors.New(resp.Error)
	}

	resultEncoded, err := getBCSResult(resp)
	if err != nil {
		return false, err
	}

	var result bool

	err = bcs.UnmarshalAll(resultEncoded[0], &result)
	if err != nil {
		return false, err
	}

	return result, nil
}

// GetLatestBlockInfo returns the block hash and height of the best block header.
func (c *LCClient) GetLatestBlockInfo(ctx context.Context) (*clients.BlockInfo, error) {
	ptb := suiptb.NewTransactionDataTransactionBuilder()

	ptb.MoveCall(
		c.PackageID,
		"light_client",
		"head",
		[]sui.TypeTag{},
		[]suiptb.CallArg{c.LcObjArg},
	)

	resp, err := c.devInspectTransactionBlock(ctx, ptb)

	if !resp.Effects.Data.IsSuccess() {
		return nil, errors.New(resp.Error)
	}

	resultEncoded, err := getBCSResult(resp)
	if err != nil {
		return nil, err
	}

	var result LightBlock

	err = bcs.UnmarshalAll(resultEncoded[0], &result)
	if err != nil {
		return nil, err
	}

	hash := result.BlockHash()
	blockInfo := &clients.BlockInfo{
		Hash:   &hash,
		Height: int64(result.Height),
	}

	return blockInfo, nil
}

// VerifySPV verifies an SPV proof against the light client's stored headers.
// TODO: finish implementation
func (c *LCClient) VerifySPV(ctx context.Context, spvProof *types.SPVProof) (int, error) {
	return 0, nil
}

// Stop performs any necessary cleanup and shutdown operations.
func (c *LCClient) Stop() {
	// TODO: Implement any necessary cleanup or shutdown logic
	fmt.Println("Stop called")
}

// moveCall is a helper function to construct and execute a Move call on the Sui blockchain.
func (c *LCClient) moveCall(
	ctx context.Context,
	function string,
	arguments []any,
) (*suiclient.SuiTransactionBlockResponse, error) {
	req := &suiclient.MoveCallRequest{
		Signer:    c.Address,
		PackageId: c.PackageID,
		Module:    "light_client",
		Function:  function,
		TypeArgs:  []string{},
		Arguments: arguments,
		GasBudget: sui.NewBigInt(10000000000),
	}

	resp, err := c.MoveCall(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("sui move call to '%s' failed: %w", function, err)
	}

	options := &suiclient.SuiTransactionBlockResponseOptions{
		ShowEffects:       true,
		ShowObjectChanges: true,
	}

	signedResp, err := c.SignAndExecuteTransaction(ctx, c.Signer, resp.TxBytes, options)
	if err != nil {
		return nil,
			fmt.Errorf("sui transaction submission for '%s' failed: %w", function, err)
	}

	// The error returned by SignAndExecuteTransactionBlock only indicates
	// whether the transaction was successfully submitted to the network.
	// It does NOT guarantee that the transaction succeeded  during execution.
	// Thats why we MUST inspect the `Effects.Status` field.
	// It will tell us about execution errors like: Abort, OutOfGas etc.
	if !signedResp.Effects.Data.IsSuccess() {
		return signedResp, fmt.Errorf("%w: function '%s' status: %s, error: %s",
			ErrSuiTransactionFailed, function, signedResp.Effects.Data.V1.Status.Status, signedResp.Effects.Data.V1.Status.Error)
	}

	return signedResp, nil
}

func (c *LCClient) devInspectTransactionBlock(
	ctx context.Context,
	ptb *suiptb.ProgrammableTransactionBuilder,
) (*suiclient.DevInspectTransactionBlockResponse, error) {
	pt := ptb.Finish()
	kind := suiptb.TransactionKind{
		ProgrammableTransaction: &pt,
	}

	txBytes, err := bcs.Marshal(kind)
	if err != nil {
		return nil, err
	}
	r := suiclient.DevInspectTransactionBlockRequest{
		SenderAddress: c.Address,
		TxKindBytes:   txBytes,
	}

	return c.DevInspectTransactionBlock(ctx, &r)
}
