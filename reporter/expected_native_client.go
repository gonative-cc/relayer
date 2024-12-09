package reporter

import (
	"context"

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	pv "github.com/cosmos/relayer/v2/relayer/provider"
	"github.com/gonative-cc/relayer/reporter/config"
	"github.com/gonative-cc/relayer/reporter/types"
)

type NativeClient interface {
	// gets the signer addr used to post txns to babylon chain
	MustGetAddr() string
	// chain level configuration (not used in relayer)
	GetConfig() *config.BabylonConfig
	// checkpoint params (k, w, checkpointTag)
	BTCCheckpointParams() (*btcctypes.QueryParamsResponse, error)
	// txn to insert bitcoin block headers to babylon chain
	InsertHeaders(ctx context.Context, msgs *types.MsgInsertHeaders) (*pv.RelayerTxResponse, error)
	// returns if given block hash is already written to babylon chain
	ContainsBTCBlock(blockHash *chainhash.Hash) (*btclctypes.QueryContainsBytesResponse, error)
	// returns the block height and hash of tip block stored in babylon chain
	BTCHeaderChainTip() (*btclctypes.QueryTipResponse, error)
	// returns the block height and hash of base (k depth down from tip?) block stored in babylon chain
	BTCBaseHeader() (*btclctypes.QueryBaseHeaderResponse, error)
	// txn to insert bitcoin spv proofs to babylon chain
	InsertBTCSpvProof(ctx context.Context, msg *types.MsgInsertBTCSpvProof) (*pv.RelayerTxResponse, error)
	Stop() error
}
