package native2ika

import (
	"context"
	"fmt"

	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/ika"
	"github.com/rs/zerolog/log"
)

// Processor handles processing transactions from the Native chain to IKA for signing.
type Processor struct {
	ikaClient    ika.Client
	db           *dal.DB
	blockFetcher BlockFetcher
}

// NewProcessor creates a new Processor instance.
func NewProcessor(ikaClient ika.Client, db *dal.DB, blockFetcher BlockFetcher) *Processor {
	return &Processor{
		ikaClient:    ikaClient,
		db:           db,
		blockFetcher: blockFetcher,
	}
}

// Run queries sign requests from Native and handles them.
func (p *Processor) Run(ctx context.Context) error {
	ikaSignRequests, err := p.db.GetPendingIkaSignRequests()
	if err != nil {
		return fmt.Errorf("error getting pending ika sign requests from db: %w", err)
	}

	if len(ikaSignRequests) == 0 {
		log.Debug().Msg("No pendning ika sign requests.")
		return nil
	}

	for _, sr := range ikaSignRequests {
		payloads := [][]byte{sr.Payload} // TODO: this wont be needed in the future when we support singing in batches
		signatures, err := p.ikaClient.ApproveAndSign(ctx, sr.DWalletID, sr.UserSig, payloads)
		if err != nil {
			return fmt.Errorf("error calling ApproveAndSign: %w", err)
		}

		err = p.db.UpdateIkaSignRequestFinalSig(sr.ID, signatures[0])
		if err != nil {
			return fmt.Errorf("error storing signature in database: %w", err)
		}

		// TODO: insert transaction to IkaTx table
	}

	return nil
}

// ProcessBlock processes a block from the Native chain and extracts the necessary data to be stored in the database.
func (p *Processor) ProcessBlock(block *tmtypes.Block) error {
	for _, event := range block.Event {
		var dwalletID string
		var userSig string
		var payload string
		for _, field := range event.fields {
			if field.Key == "dwalletID" {
				dwalletID = field.Value
			} else if field.Key == "userSig" {
				userSig = field.Value
			} else if field.Key == "payload" {
				payload = field.Value
			}
			// any other fields
		}

		signRequest := dal.IkaSignRequest{
			ID:        uint64(block.Header.Height), // Use the block height as the ID
			Payload:   payload,
			DWalletID: dwalletID,
			UserSig:   userSig,
			FinalSig:  nil,
			Timestamp: block.Header.Time.Unix(),
		}

		err := p.db.InsertIkaSignRequest(signRequest)
		if err != nil {
			return fmt.Errorf("failed to insert IkaSignRequest: %w", err)
		}
	}

	return nil
}
