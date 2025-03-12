#!/bin/bash
set -e

RELEASE_URL="https://github.com/gonative-cc/move-bitcoin-spv/archive/refs/tags/v0.1.0.tar.gz"
CONTAINER_ID="sui-node"
PACKAGE_PATH="/move-bitcoin-spv-0.1.0"

INIT_HEADERS='[0x00000030759e91f85448e42780695a7c71a6e4f4e845ecd895b19fafaeb6f5e3c030e62233287429255f254a463d90b998ba5523634da7c67ef873268e1db40d1526d5583d5b6167ffff7f2000000000,0x0000003058deb19a44c75df6d732d4dc085df09dd053c9f0db5eee57cdbfbe09fe47237776bb7462ac45b258ea7c464a19c11fef595f3e5dfbef2fc31bc94d8aefc7223c3d5b6167ffff7f2000000000,0x00000030e89c7f970db47ef7253c982270200f7009eaa3ef698d4b06c1f55848b56f24744ba0355deefd42dbd10deced2fdcf6a0f950a4f02aacd1f9fbb7efde7566d2d53d5b6167ffff7f2000000000,0x00000030792fe6e81fc1eeea11ae6a88a67060c6e8e492eeff7439168611996864119b1cace3ddc3203b8686e44d2739c45697d47c8e83b8a0e83f036b6991bf3f64ee2c3d5b6167ffff7f2002000000,0x0000003043de7b00670f41c1e92368da064553088a75374d7aac4b0a1b645658febf9e1f02ce53a61def0d99c08db78ac3d98696306fd74ece04e2a58a61ffb73dda6d963e5b6167ffff7f2001000000,0x00000030c38dec9b487eec7702a9b208cc61046e313aedfeea24192933539244d341805e0ebfd749972b2d5952585b82276afddfb22fc487f23098b98055904034170c843e5b6167ffff7f2000000000,0x00000030292e580e3b694eabbbb18b30fa22863de2de6abb7dd156c611500c801b01d845e922b7b37db1fc5a11b02998192e75a6baf7904e5b22431cd94f3ee03e93f4323e5b6167ffff7f2000000000,0x0000003010b335151fec6cd0be3fff1322e8e7b6a84ffc09682e07da040157ec0cd9d33022636b8b8cf102f3e47c2af1fd8ddceec7b46216a618d4f1af813484c031d6b13e5b6167ffff7f2000000000,0x00000030828f08644e5e78c2d99bdbcd3d0d4ba5eb10f74909b113fd8a7fa4a45febb625da3bc639c7d2c0ed61ec76f6257d4a84fe7aebeeb6c69131290c647dbefc1cad3e5b6167ffff7f2001000000,0x00000030486697206d79c9f68c60c259e9ec913c117ac6da35f44bbc57d9e4362d1ea233ed34bda2c331cb007039d7d085b08977cf21b2aad1a50a788106302d25ff79f43e5b6167ffff7f2003000000,0x0000003085bbc10dc8694fe36144c87f7737c35f9e3e8e304c61427a7cbce8b1e97004153fb8582bc04a0abb67965f6c139445bdc5d173ddc80008aa219929ab7285278f3f5b6167ffff7f2000000000,0x00000030516567e505288fe41b2fc6be9b96318c406418c7d338168fe75a26111490eb2fec401c3902aa39842e53a0c641af518957ec3aa5984a44d32e2a9f7fee2fa67a3f5b6167ffff7f2004000000]'
BTC_NETWORK=2  # regtest (0: mainnet, 1: testnet, 2: regtest)
START_HEIGHT=190



echo "Downloading and extracting light client inside the container..."
docker exec "$CONTAINER_ID" /bin/bash -c \
  "apt-get update && \
   apt-get install -y wget && \
   wget '$RELEASE_URL' -O v0.1.0.tar.gz && \
   tar -xzvf v0.1.0.tar.gz &&"


echo "Deploying light client to Sui network..."
# PUBLISH_OUTPUT="$(docker exec "$CONTAINER_ID" sui client publish --skip-dependency-verification --gas-budget 100000000 --json)" 
# docker exec "$CONTAINER_ID" sui client publish --skip-dependency-verification --gas-budget 100000000 --json
PUBLISH_OUTPUT=$(docker exec "$CONTAINER_ID" /bin/bash -c "cd '$PACKAGE_PATH' && sui client publish --skip-dependency-verification --gas-budget 100000000 --json")

PACKAGE_ID=$(echo "$PUBLISH_OUTPUT" | jq -r '.objectChanges[] | select(.type == "published") | .packageId')

if [ -z "$PACKAGE_ID" ]; then
  echo "Failed to extract Package ID!"
  echo "Publish Output:"
  echo "$PUBLISH_OUTPUT"
  exit 1
fi

echo "Package ID: $PACKAGE_ID"


echo "Initializing light client..."
# INIT_OUTPUT="$(sui client call --function new_light_client --module light_client --package "$PACKAGE_ID" --gas-budget 100000000 --args $BTC_NETWORK $START_HEIGHT "$INIT_HEADERS" 0 --json)"
INIT_OUTPUT="$(docker exec "$CONTAINER_ID" /bin/bash -c "sui client call --function new_light_client --module light_client --package '$PACKAGE_ID' --gas-budget 100000000 --args $BTC_NETWORK $START_HEIGHT \"$INIT_HEADERS\" 0 --json")"

LIGHT_CLIENT_ID=$(echo "$INIT_OUTPUT" | jq -r '.events[] | select(.type | contains("::light_client::NewLightClientEvent")) | .parsedJson.light_client_id')


# --- Error Handling ---
if [ -z "$LIGHT_CLIENT_ID" ]; then
  echo "Failed to extract Light Client ID!"
  echo "Init Output:"
  echo "$INIT_OUTPUT"
  exit 1
fi

echo "Light Client ID: $LIGHT_CLIENT_ID"

echo "Light client deployment and initialization complete."
