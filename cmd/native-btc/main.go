package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
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
		fmt.Println("Error loading .env file", err)
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Missing transaction file path")
		return
	}
	txFilePath := os.Args[1]

	txHex, err := os.ReadFile(txFilePath)
	if err != nil {
		fmt.Println("Error loading transaction file", err)
		return
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
		fmt.Println("Error creating rpc client", err)
		return
	}
	defer client.Shutdown()

	// Decode the transaction from hex
	txBytes, err := hex.DecodeString(string(txHex))
	if err != nil {
		fmt.Println("Error decoding transaction", err)
		return
	}

	var msgTx wire.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		fmt.Println("Error deserializing transaction", err)
		return
	}

	// Send the raw transaction
	txHash, err := client.SendRawTransaction(&msgTx, false)
	if err != nil {
		fmt.Println("Error sending transaction", err)
		return
	}

	fmt.Printf("Transaction broadcasted successfully. Transaction ID: %s\n", txHash.String())
}
