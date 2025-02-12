package types

import "github.com/btcsuite/btcd/wire"

type EventType int

const (
	BlockDisconnected EventType = iota
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
