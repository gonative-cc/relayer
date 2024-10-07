package native

import (
	"fmt"

	"github.com/block-vision/sui-go-sdk/sui"
)

// CreateSuiClient creates a Sui client
func CreateSuiClient(url string) (*sui.Client, error) {
	api := sui.NewSuiClient(url)
	client, ok := api.(*sui.Client)
	if !ok {
		return nil, fmt.Errorf("failed to assert type to *sui.Client")
	}
	return client, nil
}
