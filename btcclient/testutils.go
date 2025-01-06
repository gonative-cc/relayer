package btcclient

import (
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

func NewTestClientWithWsSubscriber(rpcClient *rpcclient.Client, cfg *config.BTCConfig, retrySleepTime time.Duration, maxRetrySleepTime time.Duration, blockEventChan chan *types.BlockEvent) (*Client, error) {
	net, err := GetBTCParams(cfg.NetParams)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client:            rpcClient,
		Params:            net,
		Cfg:               cfg,
		retrySleepTime:    retrySleepTime,
		maxRetrySleepTime: maxRetrySleepTime,
		blockEventChan:    blockEventChan,
	}, nil
}
