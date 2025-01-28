package native2ika

import (
	"context"
	"fmt"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/ika"
	"github.com/rs/zerolog/log"
)

// Processor handles processing transactions from the Native chain to IKA for signing.
type Processor struct {
	ikaClient ika.Client
	db        *dal.DB
}

// NewProcessor creates a new Processor instance.
func NewProcessor(ikaClient ika.Client, db *dal.DB) *Processor {
	return &Processor{
		ikaClient: ikaClient,
		db:        db,
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
		log.Info().Msg("\x1b[33mIKA signing the sign request...\x1b[0m")
		signatures, err := p.ikaClient.ApproveAndSign(ctx, sr.DWalletID, sr.UserSig, payloads)
		if err != nil {
			return fmt.Errorf("error calling ApproveAndSign: %w", err)
		}
		log.Info().Msgf("SUCCESS: IKA signed the sign request")
		err = p.db.UpdateIkaSignRequestFinalSig(sr.ID, signatures[0])
		if err != nil {
			return fmt.Errorf("error storing signature in database: %w", err)
		}

		// TODO: insert transaction to IkaTx table
	}

	return nil
}
