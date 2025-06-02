package remote2ika

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pattonkan/sui-go/sui"
	"github.com/pattonkan/sui-go/suiclient"
	"github.com/pattonkan/sui-go/suisigner"
)

// TransactionDigest is a hash of transaction encoded to a string
type TransactionDigest = string

// Client defines the methods required for interacting with the Ika network.
type Client interface {
	SignReq(
		ctx context.Context,
		dwalletCapID string,
		signMessagesID string,
		messages [][]byte,
	) (string, error)

	// TODO: we need to find out how to query
	QuerySign()
}

// client is a wrapper around the Sui client that provides functionality
// for interacting with Ika
type client struct {
	suiCl          *suiclient.ClientImpl
	Signer         *suisigner.Signer
	LcPackage      *sui.PackageId
	LcModule       string
	LcFunction     string
	DWalletPackage string
	DWalletModule  string
	GasAddr        string
	GasBudget      uint64
}

// SignOutputEventData represents the structure of the parsed JSON data
// in the SignOutputEvent.
type SignOutputEventData struct {
	Signatures [][]byte `json:"signatures"`
}

// NewClient creates a new Client instance
// `lc` Bitcoin SPV Light Client
func NewClient(
	c *suiclient.ClientImpl,
	signer *suisigner.Signer,
	spvLC SuiCtrCall,
	dwallet SuiCtrCall,
	gasAddr string,
	gasBudgetStr string,
) (Client, error) {
	lcPackage, err := sui.PackageIdFromHex(spvLC.Package)
	if err != nil {
		return nil, err
	}

	gasBudget, err := strconv.ParseUint(gasBudgetStr, 10, 64)
	if err != nil {
		return nil, err
	}

	i := &client{
		suiCl:          c,
		Signer:         signer,
		LcPackage:      lcPackage,
		LcModule:       spvLC.Module,
		LcFunction:     spvLC.Function,
		DWalletPackage: dwallet.Package,
		DWalletModule:  dwallet.Module,
		GasAddr:        gasAddr,
		GasBudget:      gasBudget,
	}
	return i, nil
}

// SignReq issues Sui transaction to request signatures for the list of messages.
// Returns transaction digest (ID).
func (c *client) SignReq(
	ctx context.Context,
	dwalletCapID string,
	signMessagesID string,
	messages [][]byte,
) (string, error) {

	// TODO: This function was only tested against dummy implementation of the dwallet module deployed locally.
	// Once it is ready, test it again
	req := &suiclient.MoveCallRequest{
		Signer:    c.Signer.Address,
		PackageId: c.LcPackage,
		Module:    c.DWalletModule,
		Function:  "approve_messages",
		TypeArgs:  []string{},
		Arguments: []interface{}{
			dwalletCapID,
			messages,
		},
		GasBudget: sui.NewBigInt(c.GasBudget),
	}
	messageApprovals, err := c.suiCl.MoveCall(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error calling approve_messages function: %w", err)
	}
	req.Function = "sign"
	req.TypeArgs = []string{}
	req.Arguments = []interface{}{
		signMessagesID,
		messageApprovals,
	}
	resp, err := c.suiCl.MoveCall(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error calling sign function: %w", err)
	}

	options := &suiclient.SuiTransactionBlockResponseOptions{
		ShowEffects:       true,
		ShowObjectChanges: true,
	}

	response, err := c.suiCl.SignAndExecuteTransaction(ctx, c.Signer, resp.TxBytes, options)
	if err != nil {
		return "", fmt.Errorf("error executing transaction block: %w", err)
	}
	return response.Digest.String(), nil

	/*
		TODO: we don't have signatures for the requested messages, we need to rework this code.

		txDigest := response.Effects.TransactionDigest
		events, err := p.c.SuiGetEvents(ctx, models.SuiGetEventsRequest{
			Digest: txDigest,
		})
		if err != nil {
			return nil, txDigest, fmt.Errorf("ika: %w", errors.Join(err, ErrEventParsing))
		}

		return extractSignatures(events[0].ParsedJson["signatures"]), txDigest, nil
	*/
}

func (c *client) QuerySign() {}

/*
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
*/
