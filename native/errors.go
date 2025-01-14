package native

import "errors"

// ErrNoFetcher is returned when the fetcher is nil.
var ErrNoFetcher = errors.New("sign requests fetcher cannot be nil")

// ErrNoNativeProcessor is returned when the nativeProcessor is nil.
var ErrNoNativeProcessor = errors.New("nativeProcessor cannot be nil")
