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

	nativeTxs, err := p.db.GetNativeTxsByStatus(dal.NativeTxStatusPending)
	if err != nil {
		return fmt.Errorf("getting unprocessed native txs from db: %w", err)
	}

	if len(nativeTxs) == 0 {
		fmt.Println("No native transactions to process.")
		return nil
	}

	for _, nativeTx := range nativeTxs {

		signatures, err := p.ikaClient.ApproveAndSign(ctx, nativeTx.DWalletCapID, nativeTx.SignMessagesID, nativeTx.Messages)
		if err != nil {
			return fmt.Errorf("error calling ApproveAndSign: %w", err)
		}

		txs, err := p.constructBtcTxsData(&nativeTx, signatures)
		if err != nil {
			return fmt.Errorf("error appending signatures: %w", err)
		}

		for _, tx := range txs {
			err = p.db.InsertTx(*tx)
			if err != nil {
				return fmt.Errorf("error storing transaction in database: %w", err)
			}
		}

		err = p.db.UpdateNativeTxStatus(nativeTx.TxID, dal.NativeTxStatusProcessed)
		if err != nil {
			return fmt.Errorf("updating native tx status: %w", err)
		}
	}

	return nil
}

// constructBtcTxsData constructs Bitcoin transactions data from Native transactions data and IKA signatures.
func (p *Processor) constructBtcTxsData(nativeTx *dal.NativeTx, signatures [][]byte) ([]*dal.Tx, error) {
	rawTxs, err := constructRawBtcTxs(nativeTx, signatures)
	if err != nil {
		return nil, err
	}

	txs := make([]*dal.Tx, 0, len(rawTxs))
	for _, rawTx := range rawTxs {
		tx := &dal.Tx{
			BtcTxID: nativeTx.TxID,
			RawTx:   rawTx,
			Hash:    []byte(""),
			Status:  dal.StatusSigned,
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

// constructRawBitcoinTransaction constructs a raw Bitcoin transaction.
// TODO: Implement the actual Bitcoin transaction construction logic.
// Most likely we will need to include the signature produced by the network
// and the tx itself should be available from native event
func constructRawBtcTxs(nativeTx *dal.NativeTx, signatures [][]byte) ([][]byte, error) {
	// Return dummy data for now
	if nativeTx == nil || signatures == nil {
		return nil, fmt.Errorf("empty")
	}
	rawTxs := make([][]byte, 0, len(signatures))

	for i := range signatures {
		rawTx := []byte(fmt.Sprintf("dummy_raw_tx_%d", i))
		rawTxs = append(rawTxs, rawTx)
	}

	return rawTxs, nil
}
