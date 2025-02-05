package lcclient

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

type Block struct {
	Hash   *chainhash.Hash
	Height int64
}

type Client struct {
	Ping                 func(int) int
	InsertHeaders        func(blockHeaders []*wire.BlockHeader) error
	GetBTCHeaderChainTip func() (Block, error)
	ContainsBTCBlock     func(blockHash *chainhash.Hash) (bool, error)
	VerifySPV            func(spvProof *types.SPVProof) (int, error)
}

func New(rpcURL string) (*Client, jsonrpc.ClientCloser, error) {
	ctx := context.Background()
	clientHandler := Client{}

	closeHandler, err := jsonrpc.NewClient(
		ctx, rpcURL, "RPCServerHandler", &clientHandler, nil,
	)
	return &clientHandler, closeHandler, err
}
