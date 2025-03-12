#!/bin/bash
set -e

make bitcoind-init
cd contrib/
echo "Starting Docker Compose services..."
docker compose up -d --wait
echo "Services started."
cd ../

./e2e/load_bitcoin_wallet.sh

echo "Services and Bitcoin wallet ready."