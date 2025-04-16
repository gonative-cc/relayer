#!/bin/bash
set -e

function get_latest_block_height_lc() {
    latest_block_hash=$(docker exec "$CONTAINER_ID" /bin/bash -c "sui client call --function latest_block_hash --module light_client --package '$PACKAGE_ID' --gas-budget 100000000 --args $LIGHT_CLIENT_ID  --dev-inspect --json")
    latest_block_height=$(echo "$latest_block_hash" | jq -r '.events[].parsedJson.height')
    echo $latest_block_height
}

function get_btc_height() {
    echo $(docker exec -it bitcoind-node bitcoin-cli getblockchaininfo | jq ".headers")
}

echo "Running E2E tests..."
CONTAINER_ID="sui-node"


echo "Start relayer to sync first time"
# Start relayer
go build ./cmd/bitcoin-spv/
./bitcoin-spv start --config ./e2e-bitcoin-spv.yml 2>stderr.log &
sleep 30
relayer_pid=$!
kill $relayer_pid
echo "Finished first sync and stop relayer"


echo "==============================="
lc_height=$(get_latest_block_height_lc)
btc_height=$(get_btc_height)

if [[ $((lc_height - btc_height)) != 0 ]]; then
    echo "Relayer not sync the btc node and lc"
    exit 1
fi

echo "Other party update light client"
echo "Insert one header to lc"
docker exec -it bitcoind-node bitcoin-cli -generate 1 > /dev/null 2>&1

new_block_header=$(docker exec bitcoind-node  /bin/bash -c "bitcoin-cli getblockheader $(docker exec bitcoind-node  /bin/bash -c "bitcoin-cli getbestblockhash") false")
docker exec "$CONTAINER_ID" /bin/bash -c "sui client call --function insert_headers --module light_client --package '$PACKAGE_ID' --gas-budget 100000000 --args $LIGHT_CLIENT_ID '[0x${new_block_header}]' --json" > /dev/null 2>&1
sleep 10
echo "Inserted header"


echo "Start relayer to sync second time"
go build ./cmd/bitcoin-spv/
./bitcoin-spv start --config ./e2e-bitcoin-spv.yml 2>stderr.log &

echo "Create more 10 block"
docker exec -it bitcoind-node bitcoin-cli -generate 10 > /dev/null 2>&1
sleep 30
relayer_pid=$!
kill $relayer_pid

echo "Finished second sync and stop relayer"


echo "Check height!"
lc_height=$(get_latest_block_height_lc)
btc_height=$(get_btc_height)

if [[ $((lc_height - btc_height)) != 0 ]]; then
    echo "Relayer not sync the btc node and lc"
    exit 1
fi

echo "Everything work well"

echo "E2E Gap handler tests completed."

exit 0
