package native

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

var rawTxBytes = []byte{
	0x02, 0x00, 0x00, 0x00, 0x01, 0xc2, 0xd3, 0x71, 0x42, 0x10, 0x82, 0xb2,
	0xe5, 0x6e, 0x4c, 0x5f, 0x8a, 0x52, 0xe3, 0xc5, 0xc4, 0x5f, 0x91, 0x14,
	0x5e, 0x8c, 0x1f, 0x14, 0x35, 0x01, 0xf3, 0x5a, 0xc6, 0xc3, 0x33, 0x84,
	0x22, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfd, 0xff, 0xff, 0xff, 0x01, 0xa0,
	0x25, 0x26, 0x00, 0x00, 0x00, 0x00, 0x00, 0x16, 0x00, 0x14, 0x0f, 0x72,
	0x53, 0x4a, 0x4b, 0x19, 0x19, 0x45, 0x93, 0x17, 0xf9, 0x99, 0xab, 0x35,
	0xe4, 0x10, 0xec, 0xa0, 0x4f, 0xe6, 0x00, 0x00, 0x00, 0x00,
}

// Mock API
func mockSelectSignReq(w http.ResponseWriter, r *http.Request) {
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
	encodedRequests, err := mockRequests.MarshalMsg(nil)
	if err != nil {
		http.Error(w, "Failed to encode MessagePack", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-msgpack")

	_, err = w.Write(encodedRequests)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// generateMockSignRequests generates mock SignRequest data dynamically.
func generateMockSignRequests(from, limit int) SignReqs {
	var requests []SignReq
	for i := from; i < from+limit; i++ {
		req := SignReq{
			//nolint: gosec // This is a mock function, and overflow is unlikely.
			ID:        uint64(i + 1),
			Payload:   rawTxBytes,
			DWalletID: fmt.Sprintf("dwallet-%d", i+1),
			UserSig:   fmt.Sprintf("user_sig-%d", i+1),
			Timestamp: time.Now().Unix(),
		}
		requests = append(requests, req)
	}
	return requests
}

// NewMockAPISignRequestFetcher creates a new APISignRequestFetcher with a mock API URL.
func NewMockAPISignRequestFetcher() (*APISignRequestFetcher, error) {
	// Create a mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(mockSelectSignReq))
	fetcher := &APISignRequestFetcher{
		APIURL: ts.URL,
	}

	return fetcher, nil
}
