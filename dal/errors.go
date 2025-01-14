package dal

import "errors"

// ErrNoDB is returned when the database is nil.
var ErrNoDB = errors.New("database cannot be nil")
