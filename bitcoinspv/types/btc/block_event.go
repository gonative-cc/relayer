package btc

import "github.com/btcsuite/btcd/wire"

// EventType represents the type of block event that occurred in the blockchain
type EventType int

const (
	// BlockDisconnected represents an event when a block is disconnected from the chain
	BlockDisconnected EventType = iota
	// BlockConnected represents an event when a new block is connected to the chain
	BlockConnected
)

// BlockEvent represents a new block event from subscription
type BlockEvent struct {
	EventType EventType
	Height    int64
	Header    *wire.BlockHeader
}

// NewBlockEvent creates and returns a new BlockEvent
func NewBlockEvent(eventType EventType, height int64, header *wire.BlockHeader) *BlockEvent {
	return &BlockEvent{
		EventType: eventType,
		Height:    height,
		Header:    header,
	}
}
