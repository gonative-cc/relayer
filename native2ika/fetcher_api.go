package native2ika

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// APISignRequestFetcher SignRequestFetcher implementation: fetches sign requests from an API.
type APISignRequestFetcher struct {
	APIURL string
}

// GetBtcSignRequests retrieves sign requests from the API.
func (f *APISignRequestFetcher) GetBtcSignRequests(from, limit int) ([]SignReq, error) {
	u, err := url.Parse(f.APIURL) // Parse the base URL
	if err != nil {
		return nil, fmt.Errorf("failed to parse API URL: %w", err)
	}
	q := u.Query()
	q.Set("from", strconv.Itoa(from))
	q.Set("limit", strconv.Itoa(limit))
	u.RawQuery = q.Encode()
	resp, err := http.Get(u.String())
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

	// The body contains marshaled SignReqs ([]SignReq)
	var requests SignReqs
	_, err = requests.UnmarshalMsg(body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SignReqs: %w", err)
	}

	return requests, nil
}
