package native

import (
	"context"
	"sync"
	"time"

	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/gonative-cc/relayer/ika"
	"github.com/rs/zerolog"
)

const (
	blocksPerMinute = 100
)

// Indexer struct responsible for calling blockchain rpc/websocket for data and
// storing that into the database.
type Indexer struct {
	logger      zerolog.Logger
	b           Blockchain
	ika         ika.Client
	lowestBlock int
}

// NewIndexer returns a new indexer struct with open connections.
func NewIndexer(ctx context.Context, b Blockchain, logger zerolog.Logger,
	startBlockHeight int, ika ika.Client) (*Indexer, error) {
	i := &Indexer{
		b:           b,
		logger:      logger.With().Str("package", "indexer").Logger(),
		lowestBlock: startBlockHeight,
		ika:         ika,
	}
	return i, i.onStart(ctx)
}

// Start indexing
func (i *Indexer) Start(ctx context.Context) error {
	newBlocks, err := i.b.SubscribeNewBlock(ctx)
	if err != nil {
		return err
	}

	oneMin := time.NewTicker(time.Second * 60)
	defer oneMin.Stop()

	for {
		select {
		case <-ctx.Done():
			return i.Close(ctx)

		case blk := <-newBlocks:
			if err := i.HandleNewBlock(ctx, blk); err != nil {
				i.logger.Err(err).Msg("error handling block")
				return err
			}

		case <-oneMin.C: // every minute. Tries to index from old blocks, if needed.
			i.logger.Info().Msgf("One minute passed")
			go i.IndexOldBlocks(ctx)
		}
	}
}

// IndexOldBlocks checks if it is needed to index old blocks and index them as needed.
func (i *Indexer) IndexOldBlocks(ctx context.Context) {
	// TODO:
	// 1. check the oldest block available
	// 2. see if we should index it

	lastBlockReceived := 0 // TODO

	heighestBlock := i.lowestBlock + blocksPerMinute
	// if the lowest block needed to index is not {IDX_BLOCKS_PER_MINUTE} behind the current
	// block, no need to try to index, wait until it is old enough.
	if heighestBlock > lastBlockReceived {
		i.logger.Info().Int("fromBlock", i.lowestBlock).Int("ToBlock", heighestBlock).Msg("no need to index old blocks")
		return
	}

	blockHeight := i.lowestBlock
	blk, minimumNodeBlkHeight, err := i.b.Block(ctx, int64(blockHeight))
	if err != nil {
		i.logger.Err(err).Int("blockHeight", blockHeight).Msg("error getting old block from blockchain")
		return
	}

	if blk == nil && minimumNodeBlkHeight != 0 {
		i.logger.Info().Int("blockHeight", blockHeight).Int("minimumNodeBlkHeight", minimumNodeBlkHeight).Msg(
			"initial block height not available on node")
		// in this case we should continue to index from the given height.
		i.lowestBlock = minimumNodeBlkHeight
		i.IndexOldBlocks(ctx)
		return
	}

	if err := i.HandleBlock(ctx, blk); err != nil {
		i.logger.Err(err).Int("blockHeight", blockHeight).Msg("error handling old block")
	}
	i.IndexBlocksFromTo(ctx, i.lowestBlock+1, heighestBlock)
}

// IndexBlocksFromTo index blocks from specific heights.
func (i *Indexer) IndexBlocksFromTo(ctx context.Context, from, to int) {
	var wg sync.WaitGroup
	mapBlockByHeight := make(map[int]*tmtypes.Block)

	for blockHeight := from; blockHeight < to; blockHeight++ {
		// TODO - check if there is anything to index, if not - early return

		i.logger.Debug().Int("blockHeight", blockHeight).Msg("indexing old block")
		wg.Add(1) // what takes a lot of time is querying blocks from node
		go func(blockHeight int) {
			defer wg.Done()
			blk, _, err := i.b.Block(ctx, int64(blockHeight))
			if err != nil {
				i.logger.Err(err).Int("blockHeight", blockHeight).Msg("error getting old block from blockchain")
				return
			}
			mapBlockByHeight[blockHeight] = blk

		}(blockHeight)
	}

	wg.Wait()
	for blockHeight := from; blockHeight < to; blockHeight++ {
		blk, ok := mapBlockByHeight[blockHeight]
		if !ok {
			continue
		}

		if err := i.HandleBlock(ctx, blk); err != nil {
			i.logger.Err(err).Int("blockHeight", blockHeight).Msg("error handling old block")
		}
	}

}

// onStart loads the starter data into blockchain.
func (i *Indexer) onStart(ctx context.Context) error {
	return i.loadChainHeader(ctx)
}

// loadChainHeader queries the chain by the last block height and sets the chain ID inside
// the blockchain structure.
func (i *Indexer) loadChainHeader(_ context.Context) error {
	chainID, height, err := i.b.ChainHeader()
	if err != nil {
		i.logger.Err(err).Msg("error loading chain header")
		return err
	}

	// TODO: load chain info state to a file or DB
	// NOTE: must be thread safe!

	i.logger.Info().Uint64("height", height).Msg("querying chainID " + chainID)
	return nil
}

// Close closes all the open connections.
func (i *Indexer) Close(ctx context.Context) error {
	return i.b.Close(ctx)
}
