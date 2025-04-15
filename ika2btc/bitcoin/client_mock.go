package bitcoin

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// MockClient Bitcoin RPC client for testing
type MockClient struct{}

// SendRawTransaction is a mock implementation that simulates sending a raw transaction.
func (m *MockClient) SendRawTransaction(_ *wire.MsgTx, _ bool) (*chainhash.Hash, error) {
	return chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
}

// GetTransaction is a mock implementation that simulates different scenarios for confirmation
func (m *MockClient) GetTransaction(txHash *chainhash.Hash) (*btcjson.GetTransactionResult, error) {
	switch txHash.String() {
	case "0000000000000000000000000000000000000000000000000000000000000001":
		return &btcjson.GetTransactionResult{
			Confirmations: 3,
		}, nil
	default:
		return &btcjson.GetTransactionResult{
			Confirmations: 6,
		}, nil
	}
}

// Shutdown is a mock implementation that is used in tests.
func (m *MockClient) Shutdown() {
}
