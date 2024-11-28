package types

import (
	"fmt"
)

const (
	// Each checkpoint is composed of two parts
	CkptNumberOfParts = 2
)

// MustNewMsgInsertBTCSpvProof returns a MsgInsertBTCSpvProof msg given the submitter address and SPV proofs of two BTC txs
func MustNewMsgInsertBTCSpvProof(submitter string, proofs []*BTCSpvProof) *MsgInsertBTCSpvProof {
	var err error
	if len(proofs) != CkptNumberOfParts {
		err = fmt.Errorf("incorrect number of proofs: want %d, got %d", CkptNumberOfParts, len(proofs))
		panic(err)
	}

	return &MsgInsertBTCSpvProof{
		Submitter: submitter,
		Proofs:    proofs,
	}
}
