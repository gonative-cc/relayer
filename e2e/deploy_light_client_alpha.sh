#!/bin/bash
# set -e

REPO_URL="https://github.com/gonative-cc/sui-bitcoin-spv"
CONTAINER_ID="sui-node"
PACKAGE_PATH="sui-bitcoin-spv/packages/bitcoin_spv"
CONFIG_FILE="e2e-bitcoin-spv.yml"

echo "Cloning light client repo"
docker exec "$CONTAINER_ID" /bin/bash -c \
       "rm -rf $PACKAGE_PATH && \
        git clone $REPO_URL "


echo "Deploying light client to Sui network..."
PUBLISH_OUTPUT=$(docker exec "$CONTAINER_ID" /bin/bash -c "cd '$PACKAGE_PATH' && sui client publish --skip-dependency-verification --gas-budget 1000000000 --json  --with-unpublished-dependencies")

# Extract Package IDs by checking for specific modules
PACKAGE_ID=$(echo "$PUBLISH_OUTPUT" | jq -r '.objectChanges[] | select(.type == "published" and (.modules | contains(["light_client"]))) | .packageId')
BTC_LIB_ID=$(echo "$PUBLISH_OUTPUT" | jq -r '.objectChanges[] | select(.type == "published" and (.modules | contains(["header"]))) | .packageId')

LOCAL_REPO_PATH="./e2e/sui-bitcoin-spv"
rm -rf "$LOCAL_REPO_PATH"
git clone "$REPO_URL" "$LOCAL_REPO_PATH"

if [ -z "$PACKAGE_ID" ] || [ -z "$BTC_LIB_ID" ]; then
  echo "Failed to extract Package IDs!"
  echo "Light Client Package ID: $PACKAGE_ID"
  echo "BTC Lib Package ID: $BTC_LIB_ID"
  echo "Publish Output:"
  echo "$PUBLISH_OUTPUT"
  exit 1
fi

echo "Light Client Package ID: $PACKAGE_ID"
echo "BTC Lib Package ID: $BTC_LIB_ID"

echo $(pwd)

sed -i "s#^SPV_PACKAGE_ID=.*#SPV_PACKAGE_ID=${PACKAGE_ID}#" ./e2e/.e2e.env
sed -i "s#^PARSER_PACKAGE_ID=.*#PARSER_PACKAGE_ID=${PACKAGE_ID}#" ./e2e/.e2e.env

cd "$LOCAL_REPO_PATH"
cp ../.e2e.env ./.env
npm i
LIGHT_CLIENT_ID=$(node scripts/new_light_client.js | jq -r '.light_client_id')
cd ../..

pwd
echo $LIGHT_CLIENT_ID

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
sed -i "s|btc_lib_pkg_id:.*|btc_lib_pkg_id: \"$BTC_LIB_ID\"|" "$CONFIG_FILE"


echo "Configuration file updated successfully."
echo "Light client deployment and initialization complete."
