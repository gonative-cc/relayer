package native

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

func TestAPISignRequestFetcher_GetBtcSignRequests(t *testing.T) {
	const mockToken = "mock-token"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+mockToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		mockSelectSignReq(w, r)
	}))
	defer ts.Close()

	fetcher := &APISignRequestFetcher{
		APIURL: ts.URL,
	}

	t.Run("Valid request", func(t *testing.T) {
		t.Setenv("NATIVE_BTCINDEXER_BEARER_TOKEN", mockToken)
		requests, err := fetcher.GetBtcSignRequests(0, 3)
		assert.NilError(t, err)
		assert.Equal(t, 3, len(requests))

		for i, req := range requests {
			assert.Equal(t, uint64(i+1), req.ID)
			assert.Equal(t, fmt.Sprintf("dwallet-%d", i+1), req.DWalletID)
		}
	})

	t.Run("Invalid API URL", func(t *testing.T) {
		t.Setenv("NATIVE_BTCINDEXER_BEARER_TOKEN", mockToken)
		fetcher.APIURL = "invalid-url"
		_, err := fetcher.GetBtcSignRequests(0, 3)
		assert.ErrorContains(t, err, "failed to make API request")
	})

	t.Run("Missing Env Var error", func(t *testing.T) {
		t.Setenv("NATIVE_BTCINDEXER_BEARER_TOKEN", "")
		_, err := fetcher.GetBtcSignRequests(0, 3)
		assert.ErrorContains(t, err, "NATIVE_BTCINDEXER_BEARER_TOKEN environment variable not set")
	})

	t.Run("API request failure (500)", func(t *testing.T) {
		t.Setenv("NATIVE_BTCINDEXER_BEARER_TOKEN", mockToken)
		failTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer failTs.Close()

		badFetcher := &APISignRequestFetcher{APIURL: failTs.URL}
		_, err := badFetcher.GetBtcSignRequests(0, 3)
		assert.ErrorContains(t, err, "status code: 500")
	})

	t.Run("API request failure", func(t *testing.T) {
		t.Setenv("NATIVE_BTCINDEXER_BEARER_TOKEN", mockToken)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		fetcher.APIURL = ts.URL
		_, err := fetcher.GetBtcSignRequests(0, 3)
		assert.Equal(t, err.Error(), "API request failed with status code: 500")
	})
}
