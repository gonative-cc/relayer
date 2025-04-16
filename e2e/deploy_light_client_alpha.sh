#!/bin/bash
# set -e

REPO_URL="https://github.com/gonative-cc/move-bitcoin-spv"
CONTAINER_ID="sui-node"
PACKAGE_PATH="move-bitcoin-spv"
CONFIG_FILE="e2e-bitcoin-spv.yml"

INIT_HEADERS='[0x0100000000000000000000000000000000000000000000000000000000000000000000003ba3edfd7a7b12b27ac72c3e67768f617fc81bc3888a51323a9fb8aa4b1e5e4adae5494dffff7f2002000000]'
BTC_NETWORK=2  # regtest (0: mainnet, 1: testnet, 2: regtest)
START_HEIGHT=0


echo "Downloading and extracting light client inside the container..."
docker exec "$CONTAINER_ID" /bin/bash -c \
  "apt-get update && \
   apt-get install -y wget git"


echo "Funding to active address"
docker exec "$CONTAINER_ID" /bin/bash -c \
       "sui client faucet"
docker exec "$CONTAINER_ID" sui client faucet --address 0x9a5779d1f633d365652efcbe3a90abf0789e6890b880af66710db5e3e3e907e1
sleep 5


echo "Download Bitcoin SPV repo"
docker exec "$CONTAINER_ID" /bin/bash -c \
       "rm -rf $PACKAGE_PATH && \
        git clone -b release/alpha $REPO_URL "

echo "Deploying light client to Sui network..."
PUBLISH_OUTPUT=$(docker exec "$CONTAINER_ID" /bin/bash -c "cd '$PACKAGE_PATH' && sui client publish --skip-dependency-verification --gas-budget 1000000000 --json")

PACKAGE_ID=$(echo "$PUBLISH_OUTPUT" | jq -r '.objectChanges[] | select(.type == "published") | .packageId')

if [ -z "$PACKAGE_ID" ]; then
  echo "Failed to extract Package ID!"
  echo "Publish Output:"
  echo "$PUBLISH_OUTPUT"
  exit 1
fi

echo "Package ID: $PACKAGE_ID"


echo "Initializing light client..."
INIT_OUTPUT="$(docker exec "$CONTAINER_ID" /bin/bash -c "sui client call --function init_light_client_network --module light_client --package '$PACKAGE_ID' --gas-budget 100000000 --args $BTC_NETWORK $START_HEIGHT \"$INIT_HEADERS\" 0 --json")"


LIGHT_CLIENT_ID=$(echo "$INIT_OUTPUT" | jq -r '.events[] | select(.type | contains("::light_client::NewLightClientEvent")) | .parsedJson.light_client_id')

if [ -z "$LIGHT_CLIENT_ID" ]; then
  echo "Failed to extract Light Client ID!"
  echo "Init Output:"
  echo "$INIT_OUTPUT"
  exit 1
fi

echo "Light Client ID: $LIGHT_CLIENT_ID"

echo "Updating $CONFIG_FILE with Package ID and Light Client ID..."

sed -i "s|lc_object_id:.*|lc_object_id: \"$LIGHT_CLIENT_ID\"|" "$CONFIG_FILE"
sed -i "s|lc_package_id:.*|lc_package_id: \"$PACKAGE_ID\"|" "$CONFIG_FILE"


echo "Configuration file updated successfully."
echo "Light client deployment and initialization complete."
