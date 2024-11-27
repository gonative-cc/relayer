package bitcoin

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// Mock Bitcoin RPC client for testing
type MockBitcoinClient struct{}

func (m *MockBitcoinClient) SendRawTransaction(_ *wire.MsgTx, _ bool) (*chainhash.Hash, error) {
	hash, err := chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (m *MockBitcoinClient) Shutdown() {
}
