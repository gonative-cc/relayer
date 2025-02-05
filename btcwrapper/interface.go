package btcwrapper

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	btcwire "github.com/btcsuite/btcd/wire"

	relayertypes "github.com/gonative-cc/relayer/bitcoinspv/types"
)

// BTCClient defines the interface for interacting with a Bitcoin node
type BTCClient interface {
	// Node lifecycle methods
	Stop()
	WaitForShutdown()

	// Block subscription methods
	SubscribeNewBlocks()
	BlockEventChannel() <-chan *relayertypes.BlockEvent

	// Block query methods
	GetBTCTipBlock() (*chainhash.Hash, int64, error)
	GetBTCBlockByHash(blockHash *chainhash.Hash) (*relayertypes.IndexedBlock, *btcwire.MsgBlock, error)
	GetBTCTailBlocksByHeight(Blockheight int64) ([]*relayertypes.IndexedBlock, error)
	GetBTCBlockByHeight(Blockheight int64) (*relayertypes.IndexedBlock, *btcwire.MsgBlock, error)
}
