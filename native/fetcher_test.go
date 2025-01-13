package native

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

func TestAPISignRequestFetcher_GetBtcSignRequests(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(mockSelectSignReq))
	defer ts.Close()

	fetcher := &APISignRequestFetcher{
		APIURL: ts.URL,
	}

	t.Run("Valid request", func(t *testing.T) {
		requests, err := fetcher.GetBtcSignRequests(0, 3)
		assert.NilError(t, err)
		assert.Equal(t, 3, len(requests))

		for i, req := range requests {
			assert.Equal(t, uint64(i+1), req.ID)
			assert.Equal(t, fmt.Sprintf("dwallet-%d", i+1), req.DWalletID)
		}
	})

	t.Run("Invalid API URL", func(t *testing.T) {
		fetcher.APIURL = "invalid-url"
		_, err := fetcher.GetBtcSignRequests(0, 3)
		assert.ErrorContains(t, err, "failed to make API request")
	})

	t.Run("API request failure", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		fetcher.APIURL = ts.URL
		_, err := fetcher.GetBtcSignRequests(0, 3)
		assert.Equal(t, err.Error(), "API request failed with status code: 500")
	})
}
