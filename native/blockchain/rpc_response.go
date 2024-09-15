package blockchain

import (
	"fmt"
)

// CheckRespID verifies if the ID sent is the same as received.
func CheckRespID(idSent, idReceived int) error {
	if idSent != idReceived {
		return fmt.Errorf("rpc ID call and response do not match. Sent: %d, Received: %d", idSent, idReceived)
	}
	return nil
}

// RPCRespChainID represents the response struct for a chain ID call.
type RPCRespChainID struct {
	ID     int `json:"id"`
	Result struct {
		Block struct {
			Header struct {
				ChainID string `json:"chain_id"`
				Height  string `json:"height"`
			} `json:"header"`
		} `json:"block"`
	} `json:"result"`
}

// CheckRespID checks if the ID sent matches.
func (r RPCRespChainID) CheckRespID(idSent int) error {
	return CheckRespID(idSent, r.ID)
}
