package suigoclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/fardream/go-bcs/bcs"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
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
	logger      zerolog.Logger
	lcPackageID *sui.PackageId
	lcObject    *suiptb.CallArg
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
		c.lcPackageID,
		"light_client",
		"exist",
		[]sui.TypeTag{},
		[]suiptb.CallArg{*c.lcObject, {Pure: &b}},
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
		c.lcPackageID,
		"light_client",
		"head",
		[]sui.TypeTag{},
		[]suiptb.CallArg{*c.lcObject},
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

	rawHeaders := make([][]byte, 0, len(blockHeaders))

	for _, header := range blockHeaders {
		rawHeader, err := toBytes(header)
		if err != nil {
			return fmt.Errorf("error serializing block header: %w", err)
		}
		rawHeaders = append(rawHeaders, rawHeader)
	}

	ptb := suiptb.NewTransactionDataTransactionBuilder()

	headers, err := bcs.Marshal(rawHeaders)
	if err != nil {
		return err
	}

	ptb.MoveCall(c.lcPackageID, "light_client", "insert_headers", []sui.TypeTag{}, []suiptb.CallArg{*c.lcObject, {Pure: &headers}})

	pt := ptb.Finish()

	txData := suiptb.NewTransactionData(c.Signer.Address, pt, nil, suiclient.DefaultGasBudget, suiclient.DefaultGasPrice)

	txBytes, err := bcs.Marshal(txData)

	if err != nil {
		return err
	}

	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		c.Signer,
		txBytes,
		&suiclient.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)

	if err != nil || !txnResponse.Effects.Data.IsSuccess() {
		return err
	}

	return nil
}
