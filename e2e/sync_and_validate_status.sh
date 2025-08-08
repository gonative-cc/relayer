#!/bin/bash
set -e
CONTAINER_ID="sui-node"
E2E_YAML_CONFIG=./e2e-bitcoin-spv.yml

LIGHT_CLIENT_ID="$(yq -r '.sui.lc_object_id' $E2E_YAML_CONFIG)"
PACKAGE_ID="$(yq -r '.sui.lc_package_id' $E2E_YAML_CONFIG)"

function get_latest_block_hash_lc() {
    latest_block_height=$(docker exec "$CONTAINER_ID" /bin/bash -c "sui client call --function head_hash --module light_client --package '$PACKAGE_ID' --gas-budget 100000000 --args $LIGHT_CLIENT_ID  --dev-inspect --json")
     echo $latest_block_height
}

function get_btc_height() {
    docker exec -i bitcoind-node bitcoin-cli getblockchaininfo | jq ".headers"
}

echo "No error" > stderr.log
# Start relayer
./out/bitcoin-spv start --config $E2E_YAML_CONFIG 2>stderr.log &
relayer_pid=$!
sleep 30
cat stderr.log
kill $relayer_pid


lc_height=$(get_latest_block_hash_lc)
btc_height=$(get_btc_height)

echo $lc_height
if [[ $((lc_height - btc_height)) != 0 ]]; then
    echo "Relayer not sync the btc node and lc"
    exit 1
fi
