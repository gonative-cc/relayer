package ika

import (
	"context"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/rs/zerolog"
)

// Client is a wrapper around the Sui client that provides functionality
// for interacting with Ika
type Client struct {
	c         *sui.Client
	Signer    *signer.Signer
	LcPackage string
	Module    string
	Function  string
	GasAddr   string
	GasBudget string
}

// NewClient creates a new Client instance
func NewClient(
	c *sui.Client,
	signer *signer.Signer,
	ctr SuiCtrCall,
	gasAddr, gasBudget string,
) (*Client, error) {
	i := &Client{
		c:         c,
		Signer:    signer,
		LcPackage: ctr.Package,
		Module:    ctr.Module,
		Function:  ctr.Function,
		GasAddr:   gasAddr,
		GasBudget: gasBudget,
	}
	return i, nil
}

// UpdateLC sends light blocks to the Native Light Client module in the Ika blockchain.
// It returns the transaction response and an error if any occurred.
func (p *Client) UpdateLC(
	ctx context.Context,
	lb *tmtypes.LightBlock,
	logger zerolog.Logger,
) (models.SuiTransactionBlockResponse, error) {
	req := models.MoveCallRequest{
		Signer:          p.Signer.Address,
		PackageObjectId: p.LcPackage,
		Module:          p.Module,
		Function:        p.Function,
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			lb,
		},
		Gas:       &p.GasAddr,
		GasBudget: p.GasBudget,
	}
	resp, err := p.c.MoveCall(ctx, req)
	if err != nil {
		logger.Err(err).Msg("Error calling move function:")
		return models.SuiTransactionBlockResponse{}, err // Return zero value for the response
	}

	// TODO: verify if we need to call this
	return p.c.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: resp,
		PriKey:      p.Signer.PriKey,
		Options: models.SuiTransactionBlockOptions{
			ShowInput:    true,
			ShowRawInput: true,
			ShowEffects:  true,
		},
		RequestType: "WaitForLocalExecution",
	})
}
