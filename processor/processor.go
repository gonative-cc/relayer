package processor

import (
	"context"
	"fmt"
	"sync"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/ika"
)

// NativeTxData represents the data extracted from a Native chain transaction.
type NativeTxData struct {
	TxID           uint64   `json:"tx_id"`
	DWalletCapID   string   `json:"dwallet_cap_id"`
	SignMessagesID string   `json:"sign_messages_id"`
	Messages       [][]byte `json:"messages"`
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

// ProcessTransaction processes a transaction from the Native chain.
func (p *Processor) ProcessTransaction(ctx context.Context, nativeTx *NativeTxData, mu *sync.Mutex) error {
	mu.Lock()
	defer mu.Unlock()

	signature, err := p.ikaClient.ApproveAndSign(ctx, nativeTx.DWalletCapID, nativeTx.SignMessagesID, nativeTx.Messages)
	if err != nil {
		return fmt.Errorf("error calling ApproveAndSign: %w", err)
	}

	tx, err := p.constructBitcoinTxData(nativeTx, signature[0])
	if err != nil {
		return fmt.Errorf("error appending signatures: %w", err)
	}

	err = p.db.InsertTx(*tx)
	if err != nil {
		return fmt.Errorf("error storing transaction in database: %w", err)
	}

	return nil
}

// constructBitcoinTxData constructs Bitcoin transaction data from Native transaction data and IKA signature.
func (p *Processor) constructBitcoinTxData(nativeTx *NativeTxData, signature []byte) (*dal.Tx, error) {
	rawTx, err := constructRawBitcoinTransaction(nativeTx, signature)
	if err != nil {
		return nil, err
	}

	tx := &dal.Tx{
		BtcTxID: nativeTx.TxID,
		RawTx:   rawTx,
		Hash:    []byte(""),
		Status:  dal.StatusSigned,
	}
	return tx, nil
}

// constructRawBitcoinTransaction constructs a raw Bitcoin transaction.
// TODO: Implement the actual Bitcoin transaction construction logic.
// Most likely we will need to include the singature produced by the network
// and the tx itself should be available from native event
func constructRawBitcoinTransaction(nativeTx *NativeTxData, signature []byte) ([]byte, error) {
	// Return dummy data for now
	println("nativeTx: %d", nativeTx)
	println("signature: %d", signature)
	return []byte("dummy_raw_bitcoin_transaction"), nil
}
