#!/bin/bash
set -e

docker exec bitcoind-node bitcoin-cli -regtest loadwallet nativewallet
echo "Bitcoin wallet loaded."