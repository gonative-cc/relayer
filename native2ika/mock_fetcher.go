package native2ika

import (
	"fmt"
	"time"
)

var rawTxBytes = []byte{
	0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x01, 0x00, 0xf2, 0x05,
	0x2a, 0x01, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x88, 0xac, 0x00, 0x00,
	0x00, 0x00,
}

var mockSignRequests = []SignRequest{
	{ID: 1, Payload: rawTxBytes, DWalletID: "dwallet1",
		UserSig: "user_sig1", FinalSig: nil, Timestamp: time.Now().Unix()},
	{ID: 2, Payload: rawTxBytes, DWalletID: "dwallet2",
		UserSig: "user_sig2", FinalSig: []byte("final_sig2"), Timestamp: time.Now().Unix()},
	{ID: 3, Payload: rawTxBytes, DWalletID: "dwallet3",
		UserSig: "user_sig3", FinalSig: nil, Timestamp: time.Now().Unix()},
	{ID: 4, Payload: rawTxBytes, DWalletID: "dwallet4",
		UserSig: "user_sig5", FinalSig: []byte("final_sig5"), Timestamp: time.Now().Unix()},
	{ID: 5, Payload: rawTxBytes, DWalletID: "dwallet4",
		UserSig: "user_sig5", FinalSig: []byte("final_sig5"), Timestamp: time.Now().Unix()},
	{ID: 6, Payload: rawTxBytes, DWalletID: "dwallet4",
		UserSig: "user_sig6", FinalSig: []byte("final_sig6"), Timestamp: time.Now().Unix()},
	{ID: 7, Payload: rawTxBytes, DWalletID: "dwallet4",
		UserSig: "user_sig7", FinalSig: []byte("final_sig7"), Timestamp: time.Now().Unix()},
	{ID: 8, Payload: rawTxBytes, DWalletID: "dwallet8",
		UserSig: "user_sig8", FinalSig: []byte("final_sig8"), Timestamp: time.Now().Unix()},
	{ID: 9, Payload: rawTxBytes, DWalletID: "dwallet9",
		UserSig: "user_sig9", FinalSig: []byte("final_sig9"), Timestamp: time.Now().Unix()},
	{ID: 10, Payload: rawTxBytes, DWalletID: "dwallet10",
		UserSig: "user_sig10", FinalSig: []byte("final_sig10"), Timestamp: time.Now().Unix()},
}

// MockSignRequestFetcher is a mock implementation of the SignRequestFetcher interface.
type MockSignRequestFetcher struct {
	SampleRequests []SignRequest
}

// FetchSignRequests returns mock sign requests.
func (m MockSignRequestFetcher) GetBtcSignRequests(from int, limit int) ([]SignRequestBytes, error) {
	var result []SignRequestBytes

	if from < 0 || from >= len(m.SampleRequests) || limit <= 0 {
		return result, nil // Return empty slice for invalid input
	}

	to := from + limit
	if to > len(m.SampleRequests) {
		to = len(m.SampleRequests)
	}

	for _, req := range m.SampleRequests[from:to] {
		packed, err := req.MarshalMsg(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal SignRequest: %w", err)
		}
		result = append(result, packed)
	}

	return result, nil
}
