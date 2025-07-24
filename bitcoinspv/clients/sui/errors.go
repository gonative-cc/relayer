package sui

import "errors"

// Errors
var (
	ErrSuiClientNil           = errors.New("sui client cannot be nil")
	ErrEmptyObjectID          = errors.New("objectID cannot be empty")
	ErrSignerNill             = errors.New("singer cannot be nil")
	ErrNoBlockHeaders         = errors.New("no block headers provided")
	ErrLightBlockHashNotFound = errors.New(
		"unexpected event data format: 'light_block_hash' field not found or not a slice",
	)
	ErrHeightNotFound       = errors.New("unexpected event data format: 'height' field not found")
	ErrHeightInvalidType    = errors.New("unexpected event data format: 'height' expected type of string")
	ErrHeightInvalidValue   = errors.New("invalid height value")
	ErrBlockHashInvalidType = errors.New("unexpected type in 'light_block_hash' array: expected float64")
	ErrBlockHashInvalidByte = errors.New("invalid byte value in 'light_block_hash' array")
	ErrBlockHashInvalid     = errors.New("invalid block hash bytes")
	ErrNoEventsFound        = errors.New("no events found for transaction digest")
	ErrEventDataFormat      = errors.New("failed to retrieve Sui events")
	ErrSuiTransactionFailed = errors.New("sui transaction execution failed")

	ErrGetObject = errors.New("sui GetObject")
)
