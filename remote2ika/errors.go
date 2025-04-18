package remote2ika

import "errors"

// ErrEventParsing is returned when there is an error parsing events.
var ErrEventParsing = errors.New("can't parse Sui event")
