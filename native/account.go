package native

import (
	"fmt"

	"github.com/block-vision/sui-go-sdk/signer"
)

// CreateSigner creates a signer
func CreateSigner(mnemonic string) (*signer.Signer, error) {
	signerAccount, err := signer.NewSignertWithMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	fmt.Printf("signerAccount.Address: %s\n", signerAccount.Address)
	return signerAccount, nil
}