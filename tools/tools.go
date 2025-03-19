//go:build tools


// This file uses the recommended method for tracking developer tools in a Go
// module.
// TODO: should be removed when updating toolchain to go v1.24 in May
//
// REF: https://go.dev/doc/modules/managing-dependencies#tools
package tools

import (
	// _ "github.com/mgechev/revive"
	_ "github.com/vektra/mockery/v2"
)
