#!/bin/bash
set -e

make bitcoind-init
cd contrib/
echo "Starting Docker Compose services..."
docker compose down
echo "Services started."
cd ../