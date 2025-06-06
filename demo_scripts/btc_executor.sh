#!/bin/bash

# This is a DEMO of the Trustless Bitcoin Execution Node, documented in:
# https://github.com/gonative-cc/sui-native/tree/master/bitcoin_executor

#running bitcoind in regtest mode required for the scirpt to work. If the wallets already exist load them instead.
bitcoin-cli -regtest createwallet "alice"
bitcoin-cli -regtest createwallet "bob"
# bitcoin-cli -regtest loadwallet "alice"
# bitcoin-cli -regtest loadwallet "bob"

ALICE_ADDRESS=$(bitcoin-cli -regtest -rpcwallet=alice getnewaddress "alice_rewards" "bech32")
echo "Alice P2WPKH address: $ALICE_ADDRESS"
BOB_ADDRESS=$(bitcoin-cli -regtest -rpcwallet=bob getnewaddress "bob_receive" "bech32")
echo "Bob P2WPKH address: $BOB_ADDRESS"

bitcoin-cli -regtest generatetoaddress 200 "$ALICE_ADDRESS"

TXID_ALICE_TO_BOB=$(bitcoin-cli -regtest -rpcwallet=alice sendtoaddress "$BOB_ADDRESS" 1.0)
echo "Tx ID Alice sending 1BTC to Bob: $TXID_ALICE_TO_BOB"

TX_HEX=$(bitcoin-cli -regtest getrawtransaction "$TXID_ALICE_TO_BOB")
echo "Tx hex: $TX_HEX"

bitcoin-cli -regtest decoderawtransaction "$TX_HEX"

# Mine one block to include the tx
bitcoin-cli -regtest generatetoaddress 1 "$ALICE_ADDRESS"
bitcoin-cli -regtest -rpcwallet=bob getbalance
