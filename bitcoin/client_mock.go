package bitcoin

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// MockClient Bitcoin RPC client for testing
type MockClient struct{}

// SendRawTransaction is a mock implementation that simulates sending a raw transaction.
func (m *MockClient) SendRawTransaction(_ *wire.MsgTx, _ bool) (*chainhash.Hash, error) {
	hash, err := chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// Shutdown is a mock implementation that does nothing.
func (m *MockClient) Shutdown() {
}
