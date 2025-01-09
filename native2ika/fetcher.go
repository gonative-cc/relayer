package native2ika

//go:generate msgp

// SignReq represents a signature request.
type SignReq struct {
	ID        uint64 `msg:"id"`
	Payload   []byte `msg:"payload"`
	DWalletID string `msg:"dwallet_id"`
	UserSig   string `msg:"user_sig"`
	FinalSig  []byte `msg:"final_sig"`
	Timestamp int64  `msg:"time"`
}

type SignReqs []SignReq

// SignRequestFetcher is an interface for getting sign requests from the Native network.
type SignReqFetcher interface {
	GetBtcSignRequests(from int, limit int) ([]SignReq, error)
}
