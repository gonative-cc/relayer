package native2ika

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

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
	GetBtcSignRequests(from int, limit int) ([]SignRequest, error)
}

// APISignRequestFetcher SignRequestFetcher implementation: fetches sign requests from an API.
type APISignRequestFetcher struct {
	APIURL string
}

// GetBtcSignRequests retrieves sign requests from the API.
func (f *APISignRequestFetcher) GetBtcSignRequests(from, limit int) ([]SignRequest, error) {
	url := fmt.Sprintf("%s/bitcoin/signrequests?from=%d&limit=%d", f.APIURL, from, limit)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// The body contains one or more marshaled SignRequests
	// TODO: i think we should not marshal it again it doesnt make sense, Instead lets return []SignRequest instead
	var requests []SignRequest
	var offset int
	for offset < len(body) {
		var req SignRequest
		var err error
		_, err = req.UnmarshalMsg(body[offset:]) // Unmarshal a single SignRequest
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to unmarshal SignRequest: %w", err)
		}

		requests = append(requests, req)
		encoded, _ := req.MarshalMsg(nil) // We need it here to caluclate the offset
		offset += len(encoded)
	}

	return requests, nil
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
