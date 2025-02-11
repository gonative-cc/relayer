package types

import (
	"errors"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// BTCSpvProof represents a Bitcoin SPV proof for verifying transaction inclusion
// Consider we have a Merkle tree with following structure:
//
//	          ROOT
//	         /    \
//	    H1234      H5555
//	   /     \       \
//	 H12     H34      H55
//	/  \    /  \     /
//
// H1  H2  H3  H4  H5
// L1  L2  L3  L4  L5
// To prove L3 was part of ROOT we need:
// - btc_transaction_index = 2 which in binary is 010
// (where 0 means going left, 1 means going right in the tree)
// - merkle_nodes we'd have H4 || H12 || H5555
// By looking at 010 we would know that H4 is a right sibling,
// H12 is left, H5555 is right again.
type BTCSpvProof struct {
	// ConfirmingBtcBlockHash is the hash of the block containing the transaction
	// Should have exactly 80 bytes
	ConfirmingBtcBlockHash chainhash.Hash

	// BtcTransaction contains the raw transaction bytes
	BtcTransaction []byte

	// BtcTransactionIndex is the index of transaction within the block
	// Index is needed to determine if currently hashed node is left or right
	BtcTransactionIndex uint32

	// MerkleNodes contains concatenated intermediate merkle tree nodes
	// Does not include root node and leaf node against which we calculate the proof
	// Each node has 32 byte length
	// Example proof: 32_bytes_of_node1 || 32_bytes_of_node2 || 32_bytes_of_node3
	// Length will always be divisible by 32
	MerkleNodes []byte
}

// ToMsgSpvProof converts a BTCSpvProof to an SPVProof message
func (btcSpvProof *BTCSpvProof) ToMsgSpvProof(txID string, txHash *chainhash.Hash) SPVProof {
	// Calculate number of merkle nodes including txHash
	numNodes := (len(btcSpvProof.MerkleNodes) / 32) + 1
	merklePath := make([]chainhash.Hash, numNodes)

	// Copy merkle nodes
	for i := 0; i < numNodes-1; i++ {
		start := i * 32
		end := (i + 1) * 32
		copy(merklePath[i][:], btcSpvProof.MerkleNodes[start:end])
	}

	// Copy txHash as last node
	lastIndex := numNodes - 1
	copy(merklePath[lastIndex][:], txHash[:])

	return SPVProof{
		BlockHash:  btcSpvProof.ConfirmingBtcBlockHash,
		TxID:       txID,
		TxIndex:    btcSpvProof.BtcTransactionIndex,
		MerklePath: merklePath,
	}
}

// MsgInsertBTCSpvProof represents a message containing BTC SPV proofs
type MsgInsertBTCSpvProof struct {
	// Submitter string
	Proofs []*BTCSpvProof
}

// SPVProof represents a simplified payment verification proof
type SPVProof struct {
	BlockHash  chainhash.Hash
	TxID       string // 32bytes hash value in string hex format
	TxIndex    uint32 // index of transaction in block
	MerklePath []chainhash.Hash
}

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

// flattenMerkleProof converts merkle proof node hashes into a single byte slice
func flattenMerkleProof(proof []*chainhash.Hash) []byte {
	var flatProof []byte
	for _, h := range proof {
		flatProof = append(flatProof, h.CloneBytes()...)
	}
	return flatProof
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
