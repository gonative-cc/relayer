package types

import (
	"errors"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func GetWrappedTxs(msg *wire.MsgBlock) []*btcutil.Tx {
	btcTxs := []*btcutil.Tx{}

	for i := range msg.Transactions {
		newTx := btcutil.NewTx(msg.Transactions[i])
		newTx.SetIndex(i)

		btcTxs = append(btcTxs, newTx)
	}

	return btcTxs
}

// createBranch takes as input flatenned representation of merkle tree i.e
// for tree:
//
//	      r
//	    /  \
//	  d1    d2
//	 /  \   / \
//	l1  l2 l3 l4
//
// slice should look like [l1, l2, l3, l4, d1, d2, r]
// also it takes number of leafs i.e nodes at lowest level of the tree and index
// of the leaf which supposed to be proven
// it returns list of hashes required to prove given index
func createBranch(nodes []*chainhash.Hash, numfLeafs uint, idx uint) []*chainhash.Hash {

	var branch []*chainhash.Hash

	// size represents number of merkle nodes at given level. At 0 level, number of
	// nodes is equal to number of leafs
	var size = numfLeafs

	var index = idx

	// i represents starting index of given level. 0 level i.e leafs always start
	// at index 0
	var i uint = 0

	for size > 1 {

		// index^1 means we want to get sibling of the node we are proving
		// ie. for index=2, index^1 = 3 and for index=3 index^1=2
		// so xoring last bit by 1, select node oposite to the node we want the proof
		// for.
		// case with `size-1` is needed when the number of leafs is not power of 2
		// and xor^1 could point outside of the tree
		j := min(index^1, size-1)

		branch = append(branch, nodes[i+j])

		// divide index by 2 as there are two times less nodes on second level
		index = index >> 1

		// after getting node at this level we move to next one by advancing i by
		// the size of the current level
		i = i + size

		// update the size to the next level size i.e (current level size / 2)
		// + 1 is needed to correctly account for cases that the last node of the level
		// is not paired.
		// example If the level is of the size 3, then next level should have size 2, not 1
		size = (size + 1) >> 1
	}

	return branch
}

// quite inefficiet method of calculating merkle proofs, created for testing purposes
func CreateProofForIdx(transactions [][]byte, idx uint) ([]*chainhash.Hash, error) {
	if len(transactions) == 0 {
		return nil, errors.New("can't calculate proof for empty transaction list")
	}

	if int(idx) >= len(transactions) {
		return nil, errors.New("provided index should be smaller that lenght of transaction list")
	}

	var txs []*btcutil.Tx
	for _, b := range transactions {
		tx, e := btcutil.NewTxFromBytes(b)

		if e != nil {
			return nil, e
		}

		txs = append(txs, tx)
	}

	store := blockchain.BuildMerkleTreeStore(txs, false)

	var storeNoNil []*chainhash.Hash

	// to correctly calculate indexes we need to filter out all nil hashes which
	// represents nodes which are empty
	for _, h := range store {
		if h != nil {
			storeNoNil = append(storeNoNil, h)
		}
	}

	branch := createBranch(storeNoNil, uint(len(transactions)), idx)

	return branch, nil
}

func SpvProofFromHeaderAndTransactions(
	headerBytes *BTCHeaderBytes,
	transactions [][]byte,
	transactionIdx uint,
) (*BTCSpvProof, error) {
	proof, e := CreateProofForIdx(transactions, transactionIdx)

	if e != nil {
		return nil, e
	}

	var flatProof []byte

	for _, h := range proof {
		flatProof = append(flatProof, h.CloneBytes()...)
	}

	spvProof := BTCSpvProof{
		ConfirmingBtcBlockHash: headerBytes.ToBlockHeader().BlockHash(),
		BtcTransaction:         transactions[transactionIdx],
		BtcTransactionIndex:    uint32(transactionIdx),
		MerkleNodes:            flatProof,
	}

	return &spvProof, nil
}
