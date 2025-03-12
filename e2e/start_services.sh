#!/bin/bash
set -e

make bitcoind-init
cd contrib/
echo "Starting Docker Compose services..."
docker compose up -d
echo "Services started."
cd ../