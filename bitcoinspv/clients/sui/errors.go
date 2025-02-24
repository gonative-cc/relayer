package sui

import "errors"

// Errors 
var (
ErrSuiClientNil = errors.New("sui client cannot be nil")
ErrEmptyObjectID = errors.New("objectID cannot be empty")
ErrSignerNill = errors.New("singer cannot be nil")
)
