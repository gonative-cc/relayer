package native

//go:generate msgp

// SignReq represents a signature request.
type SignReq struct {
	DWalletID string
	UserSig   string
	Payload   []byte
	FinalSig  []byte
	ID        uint64
	Timestamp int64
}

// SignReqs is a slice of SignReq
type SignReqs []SignReq
