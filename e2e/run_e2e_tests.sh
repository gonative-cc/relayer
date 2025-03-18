#!/bin/bash
set -e

echo "Running E2E tests..."

# Start bitcoin node and sui client
# ./e2e/start_services.sh

# sleep 10
# Setup sui devnet config
# ./e2e/init_local_net.sh

# sleep 10
# source ./e2e/deploy_light_client.sh

./e2e/autogen_block.sh 10 1 &

./e2e/bitcoin-spv.sh
echo "E2E Tests completed."
