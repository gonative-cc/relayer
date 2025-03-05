package types

import (
	"bytes"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
)

// Helper function to create a test transaction
func createTestTransaction(t *testing.T) *btcutil.Tx {
	// Create a simple test transaction
	txHex := "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff00ffffffff01000000000000000000000000000000000000000000000000000000000000000000000000"
	txBytes, err := chainhash.NewHashFromStr(txHex)
	require.NoError(t, err)
	tx, err := btcutil.NewTxFromBytes(txBytes.CloneBytes())
	require.NoError(t, err)
	return tx
}

// Helper function to serialize a transaction
func serializeTx(t *testing.T, tx *btcutil.Tx) []byte {
	var buf bytes.Buffer
	err := tx.MsgTx().Serialize(&buf)
	require.NoError(t, err)
	return buf.Bytes()
}

func TestSpvProofFromHeaderAndTransactions(t *testing.T) {
	// Create a test block header
	header := wire.NewBlockHeader(
		int32(2),           // version
		&chainhash.Hash{},  // prevBlock
		&chainhash.Hash{},  // merkleRoot
		uint32(1234567890), // timestamp
		uint32(0x1d00ffff), // bits
	)
	header.Nonce = 0
	headerBytes := btctypes.NewHeaderBytesFromBlockHeader(header)

	// Create test transactions
	tx1 := createTestTransaction(t)
	tx2 := createTestTransaction(t)
	transactions := [][]byte{
		serializeTx(t, tx1),
		serializeTx(t, tx2),
	}

	tests := []struct {
		name           string
		headerBytes    *btctypes.HeaderBytes
		transactions   [][]byte
		transactionIdx uint32
		wantErr        bool
	}{
		{
			name:           "valid proof creation",
			headerBytes:    &headerBytes,
			transactions:   transactions,
			transactionIdx: 0,
			wantErr:        false,
		},
		{
			name:           "invalid transaction index",
			headerBytes:    &headerBytes,
			transactions:   transactions,
			transactionIdx: 2,
			wantErr:        true,
		},
		{
			name:           "empty transaction list",
			headerBytes:    &headerBytes,
			transactions:   [][]byte{},
			transactionIdx: 0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proof, err := SpvProofFromHeaderAndTransactions(tt.headerBytes, tt.transactions, tt.transactionIdx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, proof)
			assert.Equal(t, tt.transactionIdx, proof.BtcTransactionIndex)
			assert.Equal(t, tt.transactions[tt.transactionIdx], proof.BtcTransaction)
		})
	}
}

func TestCreateProofForIdx(t *testing.T) {
	// Create test transactions
	tx1 := createTestTransaction(t)
	tx2 := createTestTransaction(t)
	tx3 := createTestTransaction(t)
	transactions := [][]byte{
		serializeTx(t, tx1),
		serializeTx(t, tx2),
		serializeTx(t, tx3),
	}

	tests := []struct {
		name          string
		transactions  [][]byte
		idx           uint32
		wantErr       bool
		expectedNodes int
	}{
		{
			name:          "valid proof for first transaction",
			transactions:  transactions,
			idx:           0,
			wantErr:       false,
			expectedNodes: 2, // For 3 transactions, we need 2 nodes in the proof
		},
		{
			name:          "valid proof for middle transaction",
			transactions:  transactions,
			idx:           1,
			wantErr:       false,
			expectedNodes: 2,
		},
		{
			name:          "valid proof for last transaction",
			transactions:  transactions,
			idx:           2,
			wantErr:       false,
			expectedNodes: 2,
		},
		{
			name:         "invalid index",
			transactions: transactions,
			idx:          3,
			wantErr:      true,
		},
		{
			name:         "empty transaction list",
			transactions: [][]byte{},
			idx:          0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proof, err := CreateProofForIdx(tt.transactions, tt.idx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, proof, tt.expectedNodes)
		})
	}
}

func TestToMsgSpvProof(t *testing.T) {
	// Create a test BTCSpvProof
	tx := createTestTransaction(t)
	txHash := tx.Hash()
	txID := txHash.String()

	merkleNodes := make([]byte, 32*2) // Two merkle nodes
	merkleNodes = append(merkleNodes, txHash.CloneBytes()...)

	btcSpvProof := &BTCSpvProof{
		BtcTransaction:         serializeTx(t, tx),
		MerkleNodes:            merkleNodes,
		BtcTransactionIndex:    0,
		ConfirmingBtcBlockHash: chainhash.Hash{},
	}

	msgProof := btcSpvProof.ToMsgSpvProof(txID, txHash)

	assert.Equal(t, txID, msgProof.TxID)
	assert.Equal(t, uint32(0), msgProof.TxIndex)
	assert.Equal(t, chainhash.Hash{}, msgProof.BlockHash)
	assert.Len(t, msgProof.MerklePath, 3) // 2 merkle nodes + txHash
}
