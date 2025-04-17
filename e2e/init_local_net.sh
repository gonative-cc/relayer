#!/bin/bash
set -e

CONTAINER="sui-node"

docker exec "$CONTAINER" sui client switch --address peaceful-crocidolite

ACTIVE_ADDRESS=$(docker exec "$CONTAINER" sui client active-address)
echo "Active Sui Address: $ACTIVE_ADDRESS"

ACTIVE_ENV=$(docker exec "$CONTAINER" sui client active-env)
echo "Active Sui Env: $ACTIVE_ENV"
echo "Requesting SUI from faucet..."
docker exec "$CONTAINER" sui client faucet
docker exec "$CONTAINER" sui client faucet --address 0x9a5779d1f633d365652efcbe3a90abf0789e6890b880af66710db5e3e3e907e1


sleep 5  # Wait for faucet


docker exec "$CONTAINER" sui client gas

echo "Sui network initialized, faucet request complete"
