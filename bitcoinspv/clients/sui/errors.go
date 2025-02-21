package sui

import "errors"

// ErrSuiClientNil is returned when the Sui client is nil.
var ErrSuiClientNil = errors.New("sui client cannot be nil")

// ErrEmptyObjectID is returned when the objectID is empty.
var ErrEmptyObjectID = errors.New("objectID cannot be empty")

// ErrSignerNill is returned when the signer is nil.
var ErrSignerNill = errors.New("singer cannot be nil")
