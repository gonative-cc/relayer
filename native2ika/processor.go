package native2ika

import (
	"context"
	"database/sql"
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

// Run processes pending IKA sign requests by sending them to the IKA client for signing.
// It updates the database with the results of the signing operation.
func (p *Processor) Run(ctx context.Context) error {
	ikaSignRequests, err := p.db.GetPendingIkaSignRequests(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch pending ika sign requests from db: %w", err)
	}

	if len(ikaSignRequests) == 0 {
		// no pending sr, early return
		return nil
	}

	for _, sr := range ikaSignRequests {
		payloads := [][]byte{sr.Payload} // TODO: this wont be needed in the future when we support singing in batches
		signatures, txDigest, err := p.ikaClient.ApproveAndSign(ctx, sr.DWalletID, sr.UserSig, payloads)
		if err != nil {
			if errors.Is(err, ika.ErrEventParsing) {
				ikaTx := dal.IkaTx{
					SrID:      sr.ID,
					Status:    int64(dal.Failed),
					IkaTxID:   txDigest,
					Timestamp: time.Now().Unix(),
					Note:      sql.NullString{String: "Transaction successful, but error parsing events.", Valid: true},
				}

				insertErr := p.db.InsertIkaTx(ctx, ikaTx)
				if insertErr != nil {
					return fmt.Errorf("failed inserting IkaTx: %w", insertErr)
				}
			}
			return fmt.Errorf("failed calling ApproveAndSign for srID %d: %w", sr.ID, err)
		}
		log.Info().Msgf("SUCCESS: IKA signed the sign request")
		err = p.db.UpdateIkaSignRequestFinalSig(ctx, sr.ID, signatures[0])
		if err != nil {
			return fmt.Errorf("failed to update the signature in db: %w", err)
		}

		ikaTx := dal.IkaTx{
			SrID:      sr.ID,
			Status:    int64(dal.Success),
			IkaTxID:   txDigest,
			Timestamp: time.Now().Unix(),
		}

		err = p.db.InsertIkaTx(ctx, ikaTx)
		if err != nil {
			return fmt.Errorf("failed to insert IkaTx: %w", err)
		}
	}

	return nil
}
