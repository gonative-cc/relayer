package native2ika

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

	// Marshal the entire slice
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
			DWalletID: fmt.Sprintf("dwallet%d", i+1),
			UserSig:   fmt.Sprintf("user_sig%d", i+1),
			Timestamp: time.Now().Unix(),
		}
		requests = append(requests, req)
	}
	return requests
}

// NewMockAPISignRequestFetcher creates a new APISignRequestFetcher with a mock API URL.
func NewMockAPISignRequestFetcher() (*APISignRequestFetcher, error) {
	// Create a mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(mockJSONAPI))
	fetcher := &APISignRequestFetcher{
		APIURL: ts.URL,
	}

	return fetcher, nil
}
