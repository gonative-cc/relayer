package cli

import (
	"time"

	"github.com/gonative-cc/relayer/nbtc"
)

func loadConfig() (*nbtc.RelayerConfig, error) {
	// TODO:Implement logic to load configuration from file, env variables, or flags
	return &nbtc.RelayerConfig{
		ProcessTxsInterval:    time.Second * 5,
		ConfirmTxsInterval:    time.Second * 7,
		ConfirmationThreshold: 6,
	}, nil
}
