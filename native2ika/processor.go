package native2ika

import (
	"context"
	"fmt"
	"sync"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/ika"
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

// ProcessTxs processes a transaction from the Native chain.
func (p *Processor) ProcessTxs(ctx context.Context, mu *sync.Mutex) error {
	mu.Lock()
	defer mu.Unlock()

	ikaSignRequests, err := p.db.GetPendingIkaSignRequests()
	if err != nil {
		return fmt.Errorf("error getting pending ika sign requests from db: %w", err)
	}

	if len(ikaSignRequests) == 0 {
		fmt.Println("No pendning ika sign requests.")
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