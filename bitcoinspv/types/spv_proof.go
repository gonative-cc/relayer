package types

import "github.com/btcsuite/btcd/chaincfg/chainhash"

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

// NOTE: not copied
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

// NOTE: not copied
// SPVProof represents a simplified payment verification proof
type SPVProof struct {
	BlockHash  chainhash.Hash
	TxID       string // 32bytes hash value in string hex format
	TxIndex    uint32 // index of transaction in block
	MerklePath []chainhash.Hash
}
