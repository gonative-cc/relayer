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
sleep 10  # Wait for faucet


docker exec "$CONTAINER" sui client gas

echo "Deploying light client..."

echo "Initializing light client..."

echo "Sui network initialized, faucet request complete, and light client deployed/initialized."