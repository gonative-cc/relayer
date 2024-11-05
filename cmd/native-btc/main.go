package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/joho/godotenv"
)

// ENV variables
const (
	Host = "BTC_RPC"
	User = "USER_RPC"
	Pass = "PASS_RPC"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if len(os.Args) < 2 {
		log.Fatal("Missing transaction file path")
	}
	txFilePath := os.Args[1]

	txHex, err := os.ReadFile(txFilePath)
	if err != nil {
		log.Fatal(err)
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         os.Getenv(Host),
		User:         os.Getenv(User),
		Pass:         os.Getenv(Pass),
		HTTPPostMode: true,
		DisableTLS:   false,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	// Decode the transaction from hex
	txBytes, err := hex.DecodeString(string(txHex))
	if err != nil {
		log.Fatal(err)
	}

	var msgTx wire.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		log.Fatal(err)
	}

	// Send the raw transaction
	txHash, err := client.SendRawTransaction(&msgTx, false)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Transaction broadcasted successfully. Transaction ID: %s\n", txHash.String())
}
