package types

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
)

const (
	// 1 byte for OP_RETURN opcode
	// 1 byte for OP_DATAXX, or 2 bytes for OP_PUSHDATA1 opcode
	// max 80 bytes of application specific data
	// This stems from the fact that if data in op_return is less than 75 bytes
	// one of OP_DATAXX opcodes is used (https://wiki.bitcoinsv.io/index.php/Pushdata_Opcodes#Opcodes_1-75_.280x01_-_0x4B.29)
	// but if data in op_return is between 76 and 80bytes, OP_PUSHDATA1 needs to be used
	// in which 1 byte indicates op code itself and 1 byte indicates how many bytes
	// are pushed onto stack (https://wiki.bitcoinsv.io/index.php/Pushdata_Opcodes#OP_PUSHDATA1_.2876_or_0x4C.29)
	maxOpReturnPkScriptSize = 83
)

func ExtractOpReturnData(tx *btcutil.Tx) []byte {
	msgTx := tx.MsgTx()
	opReturnData := []byte{}

	for _, output := range msgTx.TxOut {
		pkScript := output.PkScript
		pkScriptLen := len(pkScript)
		// valid op return script will have at least 2 bytes
		// - fisrt byte should be OP_RETURN marker
		// - second byte should indicate how many bytes there are in opreturn script
		if pkScriptLen > 1 &&
			pkScriptLen <= maxOpReturnPkScriptSize &&
			pkScript[0] == txscript.OP_RETURN {

			// if this is OP_PUSHDATA1, we need to drop first 3 bytes as those are related
			// to script iteslf i.e OP_RETURN + OP_PUSHDATA1 + len of bytes
			if pkScript[1] == txscript.OP_PUSHDATA1 {
				opReturnData = append(opReturnData, pkScript[3:]...)
			} else {
				// this should be one of OP_DATAXX opcodes we drop first 2 bytes
				opReturnData = append(opReturnData, pkScript[2:]...)
			}
		}
	}

	return opReturnData
}
