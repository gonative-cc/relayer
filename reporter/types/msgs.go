package types

type MsgInsertHeaders struct {
	Signer  string
	Headers []BTCHeaderBytes
}
