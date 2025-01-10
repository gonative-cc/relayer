package native

//go:generate msgp

// SignReq represents a signature request.
type SignReq struct {
	ID        uint64
	Payload   []byte
	DWalletID string
	UserSig   string
	FinalSig  []byte
	Timestamp int64
}

type SignReqs []SignReq
