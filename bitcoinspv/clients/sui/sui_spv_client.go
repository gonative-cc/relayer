package sui

import (
	"context"
	"encoding/hex"
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

const (
	insertHeadersFunc = "insert_headers"
	containsBlockFunc = "exist"
	getChainTipFunc   = "head"
	verifySPVFunc     = "verify_tx"
	lcModule          = "light_client"
	// TODO: Use better defaultGasBudget
	defaultGasBudget = 10000000000
)

// SPVClient implements the BitcoinSPV interface, interacting with a
// Bitcoin SPV light client deployed as a smart contract on Sui
type SPVClient struct {
	*suiclient.ClientImpl
	*suisigner.Signer
	PackageID *sui.PackageId
	LcObjArg  suiptb.CallArg
	logger    zerolog.Logger
}

var _ clients.BitcoinSPV = &SPVClient{}

// New BTCLIghtClientObject creates a new SPVClient instance.
// lcObjID and lcPkgID must be Sui Object ID as HEX.
func New(
	suiClient *suiclient.ClientImpl,
	signer *suisigner.Signer,
	lcObjID string,
	lcPkgID string,
	parentLogger zerolog.Logger,
) (clients.BitcoinSPV, error) {
	if suiClient == nil {
		return nil, ErrSuiClientNil
	}
	if signer == nil {
		return nil, ErrSignerNill
	}

	lcPackageID, err := sui.PackageIdFromHex(lcPkgID)
	if err != nil {
		return nil, err
	}

	lcObjectID, err := sui.ObjectIdFromHex(lcObjID)
	if err != nil {
		return nil, err
	}

	// TODO: handle this in other PR
	ctx := context.TODO()
	lcObjectResp, err := suiClient.GetObject(ctx, &suiclient.GetObjectRequest{
		ObjectId: lcObjectID,
		Options: &suiclient.SuiObjectDataOptions{
			ShowOwner: true,
		},
	})
	if err := CheckGetObjErr(lcObjID, lcObjectResp, err); err != nil {
		return nil, err
	}

	if lcObjectResp.Data.Owner.Shared == nil {
		return nil, fmt.Errorf("object '%s' is not a shared object", lcObjID)
	}

	lcObjArg := suiptb.CallArg{
		Object: &suiptb.ObjectArg{
			SharedObject: &suiptb.SharedObjectArg{
				Id:                   lcObjectResp.Data.ObjectId,
				InitialSharedVersion: *lcObjectResp.Data.Owner.Shared.InitialSharedVersion,
				Mutable:              false,
			},
		},
	}

	return &SPVClient{
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
func (c *SPVClient) InsertHeaders(ctx context.Context, blockHeaders []wire.BlockHeader) error {
	c.logger.Info().Msgf("This is block %s", c.PackageID)

	if len(blockHeaders) == 0 {
		return ErrNoBlockHeaders
	}

	headers := make([]suiptb.Argument, 0, len(blockHeaders))

	ptb := suiptb.NewTransactionDataTransactionBuilder()

	c.LcObjArg.Object.SharedObject.Mutable = true

	obj := ptb.MustObj(suiptb.ObjectArg{
		SharedObject: &suiptb.SharedObjectArg{
			Id:                   c.LcObjArg.Object.SharedObject.Id,
			InitialSharedVersion: c.LcObjArg.Object.SharedObject.InitialSharedVersion,
			Mutable:              true,
		},
	})
	for _, header := range blockHeaders {
		rawHeader, err := BlockHeaderToHex(header)
		if err != nil {
			return err
		}
		headerBytes, err := hex.DecodeString(rawHeader[2:])
		if err != nil {
			return err
		}
		headerArg, err := ptb.Pure(headerBytes)
		header := ptb.Command(suiptb.Command{
			MoveCall: &suiptb.ProgrammableMoveCall{
				Package:       c.PackageID,
				Module:        "block_header",
				Function:      "new_block_header",
				TypeArguments: []sui.TypeTag{},
				Arguments: []suiptb.Argument{
					headerArg,
				},
			},
		},
		)

		if err != nil {
			return fmt.Errorf("error serializing block header: %w", err)
		}
		headers = append(headers, header)
	}

	headerVec := ptb.Command(
		suiptb.Command{
			MakeMoveVec: &suiptb.ProgrammableMakeMoveVec{
				Type: &sui.TypeTag{Struct: &sui.StructTag{
					Address: c.PackageID,
					Module:  "block_header",
					Name:    "BlockHeader",
				}},
				Objects: headers,
			},
		},
	)

	ptb.Command(suiptb.Command{
		MoveCall: &suiptb.ProgrammableMoveCall{
			Package:       c.PackageID,
			Module:        lcModule,
			Function:      "insert_headers",
			TypeArguments: []sui.TypeTag{},
			Arguments: []suiptb.Argument{
				obj,
				headerVec,
			},
		},
	})

	return c.executeTx(ctx, ptb.Finish())

	return err
}

// ContainsBlock checks if the light client's chain includes a block with the given hash.
func (c *SPVClient) ContainsBlock(ctx context.Context, blockHash chainhash.Hash) (bool, error) {
	ptb := suiptb.NewTransactionDataTransactionBuilder()

	b, err := bcs.Marshal(blockHash[:])
	if err != nil {
		return false, err
	}

	err = ptb.MoveCall(
		c.PackageID,
		lcModule,
		containsBlockFunc,
		[]sui.TypeTag{},
		[]suiptb.CallArg{c.LcObjArg, {Pure: &b}},
	)
	if err != nil {
		return false, err
	}

	resp, err := c.devInspectTransactionBlock(ctx, ptb)
	if err != nil {
		return false, err
	}

	if !resp.Effects.Data.IsSuccess() {
		return false, fmt.Errorf("%w: function '%s' status: %s, error: %s",
			ErrSuiTransactionFailed, containsBlockFunc, resp.Effects.Data.V1.Status.Status, resp.Effects.Data.V1.Status.Error)
	}

	resultEncoded := getBCSResult(resp)
	var result bool

	err = bcs.UnmarshalAll(resultEncoded[0], &result)
	if err != nil {
		return false, err
	}

	return result, nil
}

// GetLatestBlockInfo returns the block hash and height of the best block header.
func (c *SPVClient) GetLatestBlockInfo(ctx context.Context) (*clients.BlockInfo, error) {
	ptb := suiptb.NewTransactionDataTransactionBuilder()

	err := ptb.MoveCall(
		c.PackageID,
		lcModule,
		getChainTipFunc,
		[]sui.TypeTag{},
		[]suiptb.CallArg{c.LcObjArg},
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.devInspectTransactionBlock(ctx, ptb)
	if err != nil {
		return nil, err
	}
	if !resp.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("%w: function '%s' status: %s, error: %s",
			ErrSuiTransactionFailed, getChainTipFunc, resp.Effects.Data.V1.Status.Status, resp.Effects.Data.V1.Status.Error)
	}

	resultEncoded := getBCSResult(resp)

	var result LightBlock

	err = bcs.UnmarshalAll(resultEncoded[0], &result)
	if err != nil {
		return nil, err
	}

	hash, err := result.BlockHash()
	if err != nil {
		return nil, err
	}

	// TODO: fix lint
	blockInfo := &clients.BlockInfo{
		Hash: &hash,
		// #nosec G115
		Height: int64(result.Height),
	}

	return blockInfo, nil
}

// Stop performs any necessary cleanup and shutdown operations.
func (c *SPVClient) Stop() {
	// TODO: Implement any necessary cleanup or shutdown logic
	c.logger.Info().Msg("Stop called")
}

// executionTx is a helper function to construct and execute a Move call on the Sui blockchain.
func (c *SPVClient) executeTx(
	ctx context.Context,
	pt suiptb.ProgrammableTransaction,
) (*suiclient.SuiTransactionBlockResponse, error) {
	coinPages, err := c.GetCoins(ctx, &suiclient.GetCoinsRequest{
		Owner: c.Address,
		Limit: 5,
	})
	coins := suiclient.Coins(coinPages.Data).CoinRefs()

	if err != nil {
		return nil, fmt.Errorf("fetching Sui coins failed %w", err)
	}

	tx := suiptb.NewTransactionData(c.Address, pt, coins, defaultGasBudget, suiclient.DefaultGasPrice)


	txBytes, err := bcs.Marshal(tx)

	if err != nil {
		return nil,
			fmt.Errorf("failed to serialize Sui transaction %w", err)
	}
	options := &suiclient.SuiTransactionBlockResponseOptions{
		ShowEffects:       true,
		ShowObjectChanges: true,
	}

	signedResp, err := c.SignAndExecuteTransaction(ctx, c.Signer, txBytes, options)
	if err != nil {
		return nil,
			fmt.Errorf("sui pbt transaction submission failed: %w", err)
	}

	c.logger.Info().Msgf("%s", signedResp.Effects.Data.V1.Status.Error)

	// The error returned by SignAndExecuteTransactionBlock only indicates
	// whether the transaction was successfully submitted to the network.
	// It does NOT guarantee that the transaction succeeded  during execution.
	// Thats why we MUST inspect the `Effects.Status` field.
	// It will tell us about execution errors like: Abort, OutOfGas etc.
	if !signedResp.Effects.Data.IsSuccess() {
		return signedResp, fmt.Errorf("%w: for ptb status: %s, error: %s",
			ErrSuiTransactionFailed, signedResp.Effects.Data.V1.Status.Status, signedResp.Effects.Data.V1.Status.Error)
	}

	return signedResp, nil
}

func (c *SPVClient) devInspectTransactionBlock(
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

// CheckGetObjErr inspects objectResponse struct for errors to create a proper error response.
// Sui GetObject returns error on network error. Object errors are wrapped inside the data.
// To avoid potential logic errors, we should always use this function to inspect object errors.
func CheckGetObjErr(id string, obj *suiclient.SuiObjectResponse, err error) error {
	if err != nil {
		return err
	}

	if obj.Error != nil {
		if obj.Error.Data.NotExists != nil {
			return fmt.Errorf("%w id=%s does not exist", ErrGetObject, id)
		}
		if obj.Error.Data.Deleted != nil {
			return fmt.Errorf("%w id=%s has been deleted", ErrGetObject, id)
		}
		return fmt.Errorf("%w id=%s %v", ErrGetObject, id, obj.Error)
	}

	return nil
}
