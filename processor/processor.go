package processor

import (
	"context"
	"fmt"
	"sync"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/ika"
)

// Message is an alias for []byte, representing a single message.
type Message = []byte

// NativeTxData represents the data extracted from a Native chain transaction.
type NativeTxData struct {
	TxID           uint64    `json:"tx_id"` // TODO: do we need multiple TxIDs here?
	DWalletCapID   string    `json:"dwallet_cap_id"`
	SignMessagesID string    `json:"sign_messages_id"`
	Messages       []Message `json:"messages"`
}

// Processor handles processing transactions from the Native chain to IKA for signing.
type Processor struct {
	ikaClient ika.IClient
	db        *dal.DB
}

// NewProcessor creates a new Processor instance.
func NewProcessor(ikaClient ika.IClient, db *dal.DB) *Processor {
	return &Processor{
		ikaClient: ikaClient,
		db:        db,
	}
}

// ProcessTxs processes a transaction from the Native chain.
func (p *Processor) ProcessTxs(ctx context.Context, nativeTx *NativeTxData, mu *sync.Mutex) error {
	mu.Lock()
	defer mu.Unlock()

	signatures, err := p.ikaClient.ApproveAndSign(ctx, nativeTx.DWalletCapID, nativeTx.SignMessagesID, nativeTx.Messages)
	if err != nil {
		return fmt.Errorf("error calling ApproveAndSign: %w", err)
	}

	txs, err := p.constructBtcTxsData(nativeTx, signatures)
	if err != nil {
		return fmt.Errorf("error appending signatures: %w", err)
	}

	for _, tx := range txs {
		err = p.db.InsertTx(*tx)
		if err != nil {
			return fmt.Errorf("error storing transaction in database: %w", err)
		}
	}

	return nil
}

// constructBtcTxsData constructs Bitcoin transactions data from Native transactions data and IKA signatures.
func (p *Processor) constructBtcTxsData(nativeTx *NativeTxData, signatures [][]byte) ([]*dal.Tx, error) {
	rawTxs, err := constructRawBtcTxs(nativeTx, signatures)
	if err != nil {
		return nil, err
	}

	var txs []*dal.Tx
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
func constructRawBtcTxs(nativeTx *NativeTxData, signatures [][]byte) ([][]byte, error) {
	// Return dummy data for now
	if nativeTx == nil || signatures == nil {
		return nil, fmt.Errorf("empty")
	}
	var rawTxs [][]byte

	for i := range signatures {
		rawTx := []byte(fmt.Sprintf("dummy_raw_tx_%d", i))
		rawTxs = append(rawTxs, rawTx)
	}

	return rawTxs, nil
}
