package native2ika

//go:generate msgp

// SignRequest represents a signature request.
type SignRequest struct {
	ID        uint64 `msg:"id"`
	Payload   []byte `msg:"payload"`
	DWalletID string `msg:"dwallet_id"`
	UserSig   string `msg:"user_sig"`
	FinalSig  []byte `msg:"final_sig"`
	Timestamp int64  `msg:"time"`
}

type SignRequestBytes []byte

// SignRequestFetcher is an interface for getting sign requests from the Native network.
type SignRequestFetcher interface {
	GetBtcSignRequests(from int, limit int) ([]SignRequestBytes, error)
}
