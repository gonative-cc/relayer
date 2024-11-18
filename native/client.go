package native

import (
	"context"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/rs/zerolog"
)

// IkaClient is a wrapper around the Sui client that provides functionality
// for interacting with  Ika
type IkaClient struct {
	c         *sui.Client
	Signer    *signer.Signer
	LcPackage string
	Module    string
	Function  string
	GasAddr   string
	GasBudget string
}

// NewIkaClient creates a new IkaClient instance
func NewIkaClient(
	c *sui.Client,
	signer *signer.Signer,
	lcpackage, module, function, gasAddr, gasBudget string,
) (*IkaClient, error) {
	i := &IkaClient{
		c:         c,
		Signer:    signer,
		LcPackage: lcpackage,
		Module:    module,
		Function:  function,
		GasAddr:   gasAddr,
		GasBudget: gasBudget,
	}
	return i, nil
}

// lcUpdateCall performs a light client update call on the Ika blockchain.
// It takes a context, a light block, and a logger as input.
// It returns the transaction response and an error if any occurred.
func (p *IkaClient) lcUpdateCall(
	ctx context.Context,
	lb *tmtypes.LightBlock,
	logger zerolog.Logger,
) (models.SuiTransactionBlockResponse, error) {
	rsp, err := callMoveFunction(
		ctx,
		p.c,
		p.LcPackage,
		p.Module,
		p.Function,
		p.GasBudget,
		p.Signer.Address,
		p.GasAddr,
		lb,
	)
	if err != nil {
		logger.Err(err).Msg("Error calling move function:")
		return models.SuiTransactionBlockResponse{}, err // Return zero value for the response
	}

	rsp2, err := executeTransaction(ctx, p.c, rsp, p.Signer.PriKey)
	if err != nil {
		logger.Err(err).Msg("Error executing transaction:")
		return models.SuiTransactionBlockResponse{}, err // Return zero value for the response
	}
	return rsp2, nil
}
