package native

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/tinylib/msgp/msgp"
)

// SignRequestFetcher is an interface for getting sign requests from the Native network.
type SignReqFetcher interface {
	GetBtcSignRequests(from int, limit int) ([]SignReq, error)
}

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

	var requests SignReqs
	if err = requests.DecodeMsg(msgp.NewReader(resp.Body)); err != nil {
		return nil, fmt.Errorf("failed to decode SignReqs: %w", err)
	}
	return requests, nil
}
