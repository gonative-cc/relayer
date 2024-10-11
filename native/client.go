package native

import (
	"context"
	"fmt"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/rs/zerolog"
)

// PeraClient wrapper
type PeraClient struct {
	c      *sui.Client
	Signer *signer.Signer
	// Sui package object ID
	Package   string
	Module    string
	Function  string
	GasAddr   string
	GasBudget string
}

// NewParaClient wrapper
func NewParaClient(c *sui.Client, signer *signer.Signer, Package string, Module string,
	Function string, GasAddr string, GasBudget string) (*PeraClient, error) {
	i := &PeraClient{
		c:         c,
		Signer:    signer,
		Package:   Package,
		Module:    Module,
		Function:  Function,
		GasAddr:   GasAddr,
		GasBudget: GasBudget,
	}
	return i, nil
}

func (p *PeraClient) lcUpdateCall(ctx context.Context, lb *tmtypes.LightBlock, logger zerolog.Logger) (models.SuiTransactionBlockResponse, error) {
	fmt.Printf("In lcupdatecall", p.c)
	rsp, err := callMoveFunction(ctx, p.c, p.Package, p.Module, p.Function, p.GasBudget, p.Signer.Address, p.GasAddr, lb)
	fmt.Printf("In lcupdatecall", rsp)
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
