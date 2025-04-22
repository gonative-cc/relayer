package remote2ika

import (
	"errors"
	"strings"
)

// SuiCtrCall defines a contract path to call in Sui
type SuiCtrCall struct {
	Package, Module, Function string
}

// Validate the fields of the struct
func (ctr SuiCtrCall) Validate() error {
	return errors.Join(
		checkNotEmpty("package", ctr.Package),
		checkNotEmpty("module", ctr.Module),
		checkNotEmpty("function", ctr.Function),
	)
}

func checkNotEmpty(name, s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New(name + " must be defined")
	}
	return nil
}
