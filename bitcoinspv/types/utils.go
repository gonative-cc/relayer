package types

import (
	"errors"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func GetWrappedTxs(msg *wire.MsgBlock) []*btcutil.Tx {
	btcTxs := make([]*btcutil.Tx, 0, len(msg.Transactions))

	for i, tx := range msg.Transactions {
		newTx := btcutil.NewTx(tx)
		newTx.SetIndex(i)

		btcTxs = append(btcTxs, newTx)
	}

	return btcTxs
}

// createBranch generates a merkle proof branch for a given leaf index
// Parameters:
//
//	nodes: flattened merkle tree array [leaves..intermediates..root]
//	numLeafs: number of leaf nodes in the tree
//	idx: index of leaf to generate proof for
//
// Returns: array of hashes needed to prove the leaf at idx
func createBranch(nodes []*chainhash.Hash, numLeafs uint, idx uint) []*chainhash.Hash {
	branch := make([]*chainhash.Hash, 0)
	size := numLeafs
	index := idx
	level := uint(0)

	for size > 1 {
		// Get sibling node hash
		siblingIdx := min(index^1, size-1)
		branch = append(branch, nodes[level+siblingIdx])

		// Move to parent level
		index >>= 1
		level += size
		size = (size + 1) >> 1
	}

	return branch
}

// CreateProofForIdx generates a Merkle proof for a transaction at the given index
// Returns the proof as a slice of hashes and any error encountered
func CreateProofForIdx(transactions [][]byte, idx uint32) ([]*chainhash.Hash, error) {
	// Validate inputs
	if len(transactions) == 0 {
		return nil, errors.New("can't calculate proof for empty transaction list")
	}

	if int(idx) >= len(transactions) {
		return nil, errors.New("provided index should be smaller that length of transaction list")
	}

	// Convert transaction bytes to btcutil.Tx objects
	txs := make([]*btcutil.Tx, 0, len(transactions))
	for _, txBytes := range transactions {
		tx, err := btcutil.NewTxFromBytes(txBytes)
		if err != nil {
			return nil, err
		}

		txs = append(txs, tx)
	}

	// Build Merkle tree
	merkleTree := blockchain.BuildMerkleTreeStore(txs, false)

	// Filter out nil nodes
	var validNodes []*chainhash.Hash
	for _, node := range merkleTree {
		if node != nil {
			validNodes = append(validNodes, node)
		}
	}

	// Generate proof branch
	proof := createBranch(validNodes, uint(len(transactions)), uint(idx))

	return proof, nil
}

// NOTE: modified
// SpvProofFromHeaderAndTransactions creates a simplified payment verification proof
// for a transaction at the given index using the block header and transaction list
func SpvProofFromHeaderAndTransactions(
	headerBytes *BTCHeaderBytes,
	transactions [][]byte,
	transactionIdx uint32,
) (*BTCSpvProof, error) {
	// Get merkle proof nodes for the transaction
	merkleProof, err := CreateProofForIdx(transactions, transactionIdx)
	if err != nil {
		return nil, err
	}

	// Flatten the merkle proof nodes into a single byte slice
	flattenedProof := flattenMerkleProof(merkleProof)

	// Create and return the SPV proof
	return &BTCSpvProof{
		ConfirmingBtcBlockHash: headerBytes.ToBlockHeader().BlockHash(),
		BtcTransaction:         transactions[transactionIdx],
		BtcTransactionIndex:    transactionIdx,
		MerkleNodes:            flattenedProof,
	}, nil
}

// NOTE: not copied
// flattenMerkleProof converts merkle proof node hashes into a single byte slice
func flattenMerkleProof(proof []*chainhash.Hash) []byte {
	var flatProof []byte
	for _, h := range proof {
		flatProof = append(flatProof, h.CloneBytes()...)
	}
	return flatProof
}
