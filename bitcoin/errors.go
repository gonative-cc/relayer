package bitcoin

import "errors"

// ErrNoBtcConfig is returned when the btc config is missing.
var ErrNoBtcConfig = errors.New("missing bitcoin node configuration")

// ErrNoBtcProcessor is returned when the btcProcessor is nil.
var ErrNoBtcProcessor = errors.New("btcProcessor cannot be nil")
