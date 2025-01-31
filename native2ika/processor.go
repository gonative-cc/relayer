package native2ika

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/ika"
	"github.com/rs/zerolog/log"
)

// Processor handles processing transactions from the Native chain to IKA for signing.
type Processor struct {
	ikaClient ika.Client
	db        dal.DB
}

// NewProcessor creates a new Processor instance.
func NewProcessor(ikaClient ika.Client, db dal.DB) *Processor {
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
		signatures, txDigest, err := p.ikaClient.ApproveAndSign(ctx, sr.DWalletID, sr.UserSig, payloads)
		if err != nil {
			if errors.Is(err, ika.ErrEventParsing) {
				ikaTx := dal.IkaTx{
					SrID:      sr.ID,
					Status:    dal.Failed,
					IkaTxID:   txDigest,
					Timestamp: time.Now().Unix(),
					Note:      "Transaction successful, but error parsing events.",
				}

				insertErr := p.db.InsertIkaTx(ikaTx)
				if insertErr != nil {
					return fmt.Errorf("error inserting IkaTx: %w", insertErr)
				}
			}
			return fmt.Errorf("error calling ApproveAndSign: %w", err)
		}
		log.Info().Msgf("SUCCESS: IKA signed the sign request")
		err = p.db.UpdateIkaSignRequestFinalSig(sr.ID, signatures[0])
		if err != nil {
			return fmt.Errorf("error storing signature in database: %w", err)
		}

		ikaTx := dal.IkaTx{
			SrID:      sr.ID,
			Status:    dal.Success,
			IkaTxID:   txDigest,
			Timestamp: time.Now().Unix(),
			Note:      "",
		}

		err = p.db.InsertIkaTx(ikaTx)
		if err != nil {
			return fmt.Errorf("error inserting IkaTx: %w", err)
		}
	}

	return nil
}
