package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/native/database"
)

// Config holds the configuration parameters for the Relayer
type Config struct {
	DatabasePath string `json:"databasePath"`
	BitcoinNode  string `json:"bitcoinNode"`
	// ... other config options (e.g., RPC user, password)
}

type BitcoinClient interface {
	SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error)
	Shutdown()
}

// Relayer broadcasts pending transactions from the database to the Bitcoin network.
type Relayer struct {
	config       Config
	db           *sql.DB
	btcClient    BitcoinClient
	shutdownChan chan struct{}
}

// NewRelayer creates a new Relayer instance with the given configuration.
func NewRelayer(config Config) (*Relayer, error) {
	// Initialize database connection
	err := database.InitDB(config.DatabasePath) // Call InitDB, but don't store the result
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 2. Access the database connection from the database package
	db := database.GetDB() // Get the DB connection

	// Initialize Bitcoin RPC client
	connCfg := &rpcclient.ConnConfig{
		Host:         os.Getenv("BTC_RPC"),
		User:         os.Getenv("BTC_RPC_USER"),
		Pass:         os.Getenv("BTC_RPC_PASS"),
		HTTPPostMode: true,
		DisableTLS:   false,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Bitcoin RPC client: %w", err)
	}

	return &Relayer{
		config:       config,
		db:           db,
		btcClient:    client,
		shutdownChan: make(chan struct{}),
	}, nil
}

// Start starts the relayer's main loop to broadcast transactions.
func (r *Relayer) Start() {
	// Graceful shutdown handling
	go func() {
		<-r.shutdownChan
		// Close database connection, Bitcoin client, etc.
		r.db.Close()
		r.btcClient.Shutdown()
	}()

	for {
		// 1. Get pending transactions from the database
		pendingTxs, err := r.db.GetPendingTransactions()
		if err != nil {
			// Handle error (e.g., log and retry)
			fmt.Println("Error getting pending transactions:", err)
			continue
		}

		// 2. Broadcast transactions to Bitcoin network
		for _, tx := range pendingTxs {
			txHash, err := r.btcClient.SendRawTransaction(tx.RawTx, false)
			if err != nil {
				// Handle error (e.g., log, maybe retry later)
				fmt.Println("Error broadcasting transaction:", err)
				continue
			}

			// 3. Update transaction status in the database
			err = r.db.UpdateTransactionStatus(tx.Txid, database.StatusBroadcast)
			if err != nil {
				// Handle error (e.g., log)
				fmt.Println("Error updating transaction status:", err)
			}

			// Log successful broadcast
			fmt.Println("Broadcasted transaction:", txHash)
		}

		// 4. Wait for a certain interval
		time.Sleep(time.Second * 10) // Check every 10 seconds
	}
}

// Stop initiates a graceful shutdown of the relayer.
func (r *Relayer) Stop() {
	close(r.shutdownChan)
}
