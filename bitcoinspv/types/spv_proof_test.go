package types

import (
	"bytes"
	"testing"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
)

// createTestTx creates a minimal, structurally valid Bitcoin transaction.
func createTestTx(t *testing.T) *wire.MsgTx {
	t.Helper()
	tx := wire.NewMsgTx(wire.TxVersion)
	prevHash := chainhash.Hash{}
	prevOutPoint := wire.NewOutPoint(&prevHash, 0)
	txIn := wire.NewTxIn(prevOutPoint, []byte{0x00, 0x01}, nil)
	tx.AddTxIn(txIn)
	txOut := wire.NewTxOut(10000, []byte{0x00, 0x14, 0x75, 0x1e, 0x76, 0xe8, 0x19, 0x91, 0x96, 0xd4, 0x54, 0x94, 0x1c, 0x45, 0xd1, 0xb3, 0xa3, 0x23, 0xf1, 0x43, 0x3b, 0xd6})
	tx.AddTxOut(txOut)
	return tx
}

// serializeTx serializes a transaction to bytes.
func serializeTx(t *testing.T, tx *wire.MsgTx) []byte {
	t.Helper()
	var buf bytes.Buffer
	err := tx.Serialize(&buf)
	require.NoError(t, err)
	return buf.Bytes()
}

func TestSpvProofFromHeaderAndTransactions(t *testing.T) {
	tx1, tx2 := createTestTx(t), createTestTx(t)
	txs := [][]byte{serializeTx(t, tx1), serializeTx(t, tx2)}

	btcutilTxs := []*btcutil.Tx{btcutil.NewTx(tx1), btcutil.NewTx(tx2)}
	merkleRoot := blockchain.BuildMerkleTreeStore(btcutilTxs, false)[len(btcutilTxs)-1]

	header := wire.NewBlockHeader(2, &chainhash.Hash{}, merkleRoot, 1234567890, 0x1d00ffff)
	header.Nonce = 0
	headerBytes := btctypes.NewHeaderBytesFromBlockHeader(header)

	tests := []struct {
		name        string
		headerBytes *btctypes.HeaderBytes
		txs         [][]byte
		txID        uint32
		expectedErr error
	}{
		{
			name:        "valid proof creation",
			headerBytes: &headerBytes,
			txs:         txs,
			txID:        0,
			expectedErr: nil,
		},
		{
			name:        "invalid transaction index",
			headerBytes: &headerBytes,
			txs:         txs,
			txID:        2,
			expectedErr: errIndexOutOfBounds,
		},
		{
			name:        "empty transaction list",
			headerBytes: &headerBytes,
			txs:         [][]byte{},
			txID:        0,
			expectedErr: errEmptyTxList,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proof, err := SpvProofFromHeaderAndTransactions(tt.headerBytes, tt.txs, tt.txID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, proof)
			require.Equal(t, tt.txID, proof.BtcTransactionIndex)
			require.Equal(t, tt.txs[tt.txID], proof.BtcTransaction)
			require.Equal(t, headerBytes.ToBlockHeader().BlockHash(), proof.ConfirmingBtcBlockHash)

		})
	}
}

func TestCreateProofForIdx(t *testing.T) {
	tx1, tx2, tx3 := createTestTx(t), createTestTx(t), createTestTx(t)
	txs := [][]byte{serializeTx(t, tx1), serializeTx(t, tx2), serializeTx(t, tx3)}

	tests := []struct {
		name          string
		txs           [][]byte
		txID          uint32
		expectedErr   error
		expectedNodes int
	}{
		{
			name:          "valid proof for first transaction",
			txs:           txs,
			txID:          0,
			expectedErr:   nil,
			expectedNodes: 2, // For 3 transactions, we need 2 nodes in the proof
		},
		{
			name:          "valid proof for middle transaction",
			txs:           txs,
			txID:          1,
			expectedErr:   nil,
			expectedNodes: 2,
		},
		{
			name:          "valid proof for last transaction",
			txs:           txs,
			txID:          2,
			expectedErr:   nil,
			expectedNodes: 2,
		},
		{
			name:        "invalid index",
			txs:         txs,
			txID:        3,
			expectedErr: errIndexOutOfBounds,
		},
		{
			name:        "empty transaction list",
			txs:         [][]byte{},
			txID:        0,
			expectedErr: errEmptyTxList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proof, err := CreateProofForIdx(tt.txs, tt.txID)
			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, proof)
			assert.Len(t, proof, tt.expectedNodes)
		})
	}
}
