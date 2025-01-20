package types

import "github.com/btcsuite/btcd/chaincfg/chainhash"

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
	// Should have exactly 80 bytes
	ConfirmingBtcBlockHash chainhash.Hash
	BtcTransaction         []byte
	// Index of transaction within the block. Index is needed to determine if
	// currently hashed node is left or right.
	BtcTransactionIndex uint32
	// List of concatenated intermediate merkle tree nodes, without root node and
	// leaf node against which we calculate the proof. Each node has 32 byte
	// length. Example proof can look like: 32_bytes_of_node1 || 32_bytes_of_node2
	// ||  32_bytes_of_node3 so the length of the proof will always be divisible
	// by 32.
	MerkleNodes []byte
}

// NOTE: not copied
func (btcSpvProof *BTCSpvProof) ToMsgSpvProof(txID string, txHash *chainhash.Hash) SPVProof {
	merklePath := make([]chainhash.Hash, (len(btcSpvProof.MerkleNodes)/32)+1)
	for i := 0; i < len(btcSpvProof.MerkleNodes)/32; i++ {
		copy(merklePath[i][:], btcSpvProof.MerkleNodes[i*32:(i+1)*32])
	}
	// copy txHash to end of merklePath
	copy(merklePath[len(merklePath)-1][:], txHash[:])

	return SPVProof{
		BlockHash:  btcSpvProof.ConfirmingBtcBlockHash,
		TxID:       txID,
		TxIndex:    btcSpvProof.BtcTransactionIndex,
		MerklePath: merklePath,
	}
}

type MsgInsertBTCSpvProof struct {
	// Submitter string
	Proofs []*BTCSpvProof
}

// NOTE: not copied
type SPVProof struct {
	BlockHash  chainhash.Hash
	TxID       string // 32bytes hash value in string hex format
	TxIndex    uint32 // index of transaction in block
	MerklePath []chainhash.Hash
}
