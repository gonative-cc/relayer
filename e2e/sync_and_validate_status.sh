#!/bin/bash
set -e
CONTAINER_ID="sui-node"

function get_latest_block_height_lc() {
    latest_block_hash=$(docker exec "$CONTAINER_ID" /bin/bash -c "sui client call --function latest_block_hash --module light_client --package '$PACKAGE_ID' --gas-budget 100000000 --args $LIGHT_CLIENT_ID  --dev-inspect --json")
    latest_block_height=$(echo "$latest_block_hash" | jq -r '.events[].parsedJson.height')
    echo $latest_block_height
}

function get_btc_height() {
    echo $(docker exec -it bitcoind-node bitcoin-cli getblockchaininfo | jq ".headers")
}

# Start relayer
./out/bitcoin-spv start --config ./e2e-bitcoin-spv.yml > /dev/null 2>&1 &
sleep 30
relayer_pid=$!
kill $relayer_pid

lc_height=$(get_latest_block_height_lc)
btc_height=$(get_btc_height)
if [[ $((lc_height - btc_height)) != 0 ]]; then
    echo "Relayer not sync the btc node and lc"
    exit 1
fi
