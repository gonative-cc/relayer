package errors

import "errors"

// ErrNoDB is returned when the database is nil.
var ErrNoDB = errors.New("database cannot be nil")

// ErrNoBtcConfig is returned when the btc config is missing.
var ErrNoBtcConfig = errors.New("missing bitcoin node configuration")

// ErrNoNativeProcessor is returned when the nativeProcessor is nil.
var ErrNoNativeProcessor = errors.New("nativeProcessor cannot be nil")

// ErrNoNativeProcessor is returned when the btcProcessor is nil.
var ErrNoBtcProcessor = errors.New("btcProcessor cannot be nil")

// ErrNoFetcher is returned when the fetcher is nil.
var ErrNoFetcher = errors.New("sign requests fetcher cannot be nil")
