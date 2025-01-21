package ika

import (
	"context"
	"fmt"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Client defines the methods required for interacting with the Ika network.
type Client interface {
	UpdateLC(
		ctx context.Context,
		lb *tmtypes.LightBlock,
		logger zerolog.Logger,
	) (models.SuiTransactionBlockResponse, error)
	ApproveAndSign(ctx context.Context, dwalletCapID, signMessagesID string, messages [][]byte) ([][]byte, error)
}

// client is a wrapper around the Sui client that provides functionality
// for interacting with Ika
type client struct {
	c              *sui.Client
	Signer         *signer.Signer
	LcPackage      string
	LcModule       string
	LcFunction     string
	DWalletPackage string
	DWalletModule  string
	GasAddr        string
	GasBudget      string
}

// SignOutputEventData represents the structure of the parsed JSON data
// in the SignOutputEvent.
type SignOutputEventData struct {
	Signatures [][]byte `json:"signatures"`
}

// NewClient creates a new Client instance
func NewClient(
	c *sui.Client,
	signer *signer.Signer,
	ctr SuiCtrCall,
	dwallet SuiCtrCall,
	gasAddr, gasBudget string,
) (Client, error) {
	i := &client{
		c:              c,
		Signer:         signer,
		LcPackage:      ctr.Package,
		LcModule:       ctr.Module,
		LcFunction:     ctr.Function,
		DWalletPackage: dwallet.Package,
		DWalletModule:  dwallet.Module,
		GasAddr:        gasAddr,
		GasBudget:      gasBudget,
	}
	return i, nil
}

// UpdateLC sends light blocks to the Native Light Client module in the Ika blockchain.
// It returns the transaction response and an error if any occurred.
func (p *client) UpdateLC(
	ctx context.Context,
	lb *tmtypes.LightBlock,
	logger zerolog.Logger,
) (models.SuiTransactionBlockResponse, error) {
	req := models.MoveCallRequest{
		Signer:          p.Signer.Address,
		PackageObjectId: p.LcPackage,
		Module:          p.LcModule,
		Function:        p.LcFunction,
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

// ApproveAndSign approves and signs a set of messages using the IKA network. Returns its signatures
func (p *client) ApproveAndSign(
	ctx context.Context,
	dwalletCapID string,
	signMessagesID string,
	messages [][]byte,
) ([][]byte, error) {

	// TODO: This function was only tested against dummy implementation of the dwallet module deployed locally.
	// Once it is ready, test it again
	req := models.MoveCallRequest{
		Signer:          p.Signer.Address,
		PackageObjectId: p.DWalletPackage,
		Module:          p.DWalletModule,
		Function:        "approve_messages",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			"0x2",
			messages,
		},
		GasBudget: p.GasBudget,
	}
	log.Debug().Msg(fmt.Sprintf("signer %s", p.Signer.Address))
	_, err := p.c.MoveCall(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error calling approve_messages function: %w", err)
	}
	objectIDs := []string{
		"0xa5b4007bd478fbb155ade6a94501788ab13300904590fc33dbd777a1770cf483",
	}

	req.Function = "sign"
	req.TypeArguments = []interface{}{}
	req.Arguments = []interface{}{
		"0xa5b4007bd478fbb155ade6a94501788ab13300904590fc33dbd777a1770cf483",
		objectIDs,
	}
	resp, err := p.c.MoveCall(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error calling sign function: %w", err)
	}

	response, err := p.c.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: resp,
		PriKey:      p.Signer.PriKey,
		Options: models.SuiTransactionBlockOptions{
			ShowInput:    true,
			ShowRawInput: true,
			ShowEffects:  true,
		},
		RequestType: "WaitForLocalExecution",
	})
	if err != nil {
		return nil, fmt.Errorf("error executing transaction block: %w", err)
	}

	events, err := p.c.SuiGetEvents(ctx, models.SuiGetEventsRequest{
		Digest: response.Effects.TransactionDigest,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting events: %w", err)
	}

	return extractSignatures(events[0].ParsedJson["signatures"]), nil
}

// extractSignatures extracts bytes from the `ParsedJson` structure
func extractSignatures(data interface{}) [][]byte {
	var byteArrays [][]byte

	if slice, ok := data.([]interface{}); ok {
		for _, inner := range slice {
			var byteArray []byte
			if innerSlice, ok := inner.([]interface{}); ok {
				for _, value := range innerSlice {
					if num, ok := value.(float64); ok {
						byteArray = append(byteArray, byte(int(num)))
					}
				}
			}
			byteArrays = append(byteArrays, byteArray)
		}
	}

	return byteArrays
}
