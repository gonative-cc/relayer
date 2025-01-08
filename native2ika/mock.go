package native2ika

import (
	"fmt"
	"net/http"
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

// Mock JSON API handler
func mockJSONAPI(w http.ResponseWriter, r *http.Request) {
	// Get the 'from' and 'limit' query parameters
	from := 0
	limit := 5 // Default limit
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if _, err := fmt.Sscan(fromStr, &from); err != nil {
			http.Error(w, "Invalid 'from' parameter", http.StatusBadRequest)
			return
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if _, err := fmt.Sscan(limitStr, &limit); err != nil {
			http.Error(w, "Invalid 'limit' parameter", http.StatusBadRequest)
			return
		}
	}

	mockRequests := generateMockSignRequests(from, limit)

	// Marshal the data
	encodedRequests := make([]SignRequestBytes, 0, len(mockRequests))
	for _, req := range mockRequests {
		packed, err := req.MarshalMsg(nil)
		if err != nil {
			http.Error(w, "Failed to encode MessagePack", http.StatusInternalServerError)
			return
		}
		encodedRequests = append(encodedRequests, packed)
	}

	w.Header().Set("Content-Type", "application/x-msgpack")

	for _, encoded := range encodedRequests {
		_, err := w.Write(encoded)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}
}

// generateMockSignRequests generates mock SignRequest data dynamically.
func generateMockSignRequests(from, limit int) []SignRequest {
	var requests []SignRequest
	for i := from; i < from+limit; i++ {
		if i+1 < 0 {
			panic(fmt.Sprintf("integer overflow: i+1 = %d", i+1))
		}
		req := SignRequest{
			ID:        uint64(i + 1),
			Payload:   rawTxBytes,
			DWalletID: fmt.Sprintf("dwallet%d", i+1),
			UserSig:   fmt.Sprintf("user_sig%d", i+1),
			Timestamp: time.Now().Unix(),
		}
		requests = append(requests, req)
	}
	return requests
}
