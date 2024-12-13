package lcclient

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/filecoin-project/go-jsonrpc"
)

type Client struct {
	Ping                 func(int) int
	InsertHeaders        func(blockHeaders []*wire.BlockHeader) error
	GetBTCHeaderChainTip func() (*chainhash.Hash, error)
}

func New(rpcUrl string) (*Client, jsonrpc.ClientCloser, error) {
	ctx := context.Background()
	clientHandler := Client{}

	closeHandler, err := jsonrpc.NewClient(
		ctx, rpcUrl, "", &clientHandler, nil,
	)
	return &Client{}, closeHandler, err
}
