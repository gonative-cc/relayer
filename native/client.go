package native

import (
	"context"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/rs/zerolog"
)

// PeraClient wrapper
type PeraClient struct {
	c         *sui.Client
	Signer    *signer.Signer
	LcPackage string
	Module    string
	Function  string
	GasAddr   string
	GasBudget string
}

// NewParaClient creates a new PeraClient instance
func NewParaClient(c *sui.Client, signer *signer.Signer, lcpackage string, 
	module string, function string, gasAddr string, gasBudget string) (*PeraClient, error) {
	i := &PeraClient{
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

func (p *PeraClient) lcUpdateCall(ctx context.Context, lb *tmtypes.LightBlock, 
	logger zerolog.Logger) (models.SuiTransactionBlockResponse, error) {
	rsp, err := callMoveFunction(ctx, p.c, p.LcPackage, p.Module, p.Function, p.GasBudget, p.Signer.Address, p.GasAddr, lb)
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
