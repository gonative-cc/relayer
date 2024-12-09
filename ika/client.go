package ika

import (
	"context"
	"fmt"

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

// ApproveAndSign approves and signs a set of messages using the IKA network. Returns its signatures
func (p *Client) ApproveAndSign(
	ctx context.Context,
	dwalletCapID string,
	signMessagesID string,
	messages [][]byte,
	logger zerolog.Logger,
) ([][]byte, error) {

	//TODO: This function was only tested against dummy implementation of the dwallet module deployed locally.
	// Once it is ready, test it again
	req := models.MoveCallRequest{
		Signer:          p.Signer.Address,
		PackageObjectId: p.LcPackage,
		Module:          p.Module,
		Function:        "approve_messages",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			dwalletCapID,
			messages,
		},
		GasBudget: p.GasBudget,
	}
	messageApprovals, err := p.c.MoveCall(ctx, req)
	if err != nil {
		logger.Err(err).Msg("Error calling move function:")
		return nil, err
	}

	req.Function = "sign"
	req.TypeArguments = []interface{}
	req.Argumetns = []interface{}{
		signMessagesID,
		messageApprovals,
	}
	resp, err := p.c.MoveCall(ctx, req)
	if err != nil {
		logger.Err(err).Msg("Error calling move function:")
		return nil, err
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
		logger.Err(err).Msg("Error executing transaction block:")
		return nil, err
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
