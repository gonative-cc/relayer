package suigoclient

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

// LCClient is Bitcoin Light Client
type LCClient struct {
	*suiclient.ClientImpl
	*suisigner.Signer
	logger    zerolog.Logger
	PackageID *sui.PackageId
	LcObject  *suiclient.SuiObjectData
}

var _ clients.BitcoinSPV = &LCClient{}

// New LC client
func New(suiClient *suiclient.ClientImpl, signer *suisigner.Signer, lightClientObjectIDHex string, lightClientPackageIDHex string, parentLogger zerolog.Logger) (clients.BitcoinSPV, error) {
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
	if err != nil {
		return nil, err
	}

	return &LCClient{
		ClientImpl: suiClient,
		Signer:     signer,
		logger:     configureClientLogger(parentLogger),
		PackageID:  lcPackageID,
		LcObject:   lcObjectResp.Data,
	}, nil
}

func (c *LCClient) lcObjMut() suiptb.CallArg {
	return suiptb.CallArg{
		Object: &suiptb.ObjectArg{
			SharedObject: &suiptb.SharedObjectArg{
				Id:                   c.LcObject.ObjectId,
				InitialSharedVersion: *c.LcObject.Owner.Shared.InitialSharedVersion,
				Mutable:              true,
			},
		},
	}
}

func (c *LCClient) lcObjImmu() suiptb.CallArg {
	return suiptb.CallArg{
		Object: &suiptb.ObjectArg{
			SharedObject: &suiptb.SharedObjectArg{
				Id:                   c.LcObject.ObjectId,
				InitialSharedVersion: *c.LcObject.Owner.Shared.InitialSharedVersion,
				Mutable:              false,
			},
		},
	}
}

func configureClientLogger(parentLogger zerolog.Logger) zerolog.Logger {
	return parentLogger.With().Str("module", "spv_client").Logger()
}

func (c *LCClient) devInspectTransactionBlock(ctx context.Context, ptb *suiptb.ProgrammableTransactionBuilder) (*suiclient.DevInspectTransactionBlockResponse, error) {
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

// ContainsBlock check block hash exist in light client
func (c *LCClient) ContainsBlock(ctx context.Context, blockHash chainhash.Hash) (bool, error) {
	ptb := suiptb.NewTransactionDataTransactionBuilder()

	b, err := bcs.Marshal(blockHash[:])
	if err != nil {
		return false, err
	}

	ptb.MoveCall(
		c.PackageID,
		"light_client",
		"exist",
		[]sui.TypeTag{},
		[]suiptb.CallArg{c.lcObjImmu(), {Pure: &b}},
	)

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

// GetLatestBlockInfo check block hash exist in light client
func (c *LCClient) GetLatestBlockInfo(ctx context.Context) (*clients.BlockInfo, error) {
	ptb := suiptb.NewTransactionDataTransactionBuilder()

	ptb.MoveCall(
		c.PackageID,
		"light_client",
		"head",
		[]sui.TypeTag{},
		[]suiptb.CallArg{c.lcObjImmu()},
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

// InsertHeaders insert header
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
		c.lcObjImmu().Object.SharedObject.Id,
		rawHeaders,
	}

	c.logger.Debug().Msgf("Calling insert headers with the following arguemts: %v", arguments...)

	_, err := c.moveCall(ctx, "insert_headers", arguments)
	return err
}

// VerifySPV verifies an SPV proof against the light client's stored headers.
// TODO: finish implementation
func (c *LCClient) VerifySPV(_ context.Context, _ *types.SPVProof) (int, error) {
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
		Signer:    c.Signer.Address,
		PackageId: c.PackageID,
		Module:    "light_client",
		Function:  function,
		TypeArgs:  []string{},
		Arguments: arguments,
		GasBudget: sui.NewBigInt(suiclient.DefaultGasBudget),
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
