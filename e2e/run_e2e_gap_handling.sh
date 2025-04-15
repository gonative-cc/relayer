#!/bin/bash
# set -e

echo "Running E2E tests..."
CONTAINER_ID="sui-node"


docker exec "$CONTAINER_ID" /bin/bash -c "sui client call --function latest_block_hash --module light_client --package '$PACKAGE_ID' --gas-budget 100000000 --args $LIGHT_CLIENT_ID --json"

echo $PACKAGE_ID
# Start relayer
go build ./cmd/bitcoin-spv/
./bitcoin-spv start --config ./e2e-bitcoin-spv.yml 2>stderr.log &
sleep 10


new_block_header=$(docker exec bitcoind-node  /bin/bash -c "bitcoin-cli getblockheader $(docker exec bitcoind-node  /bin/bash -c "bitcoin-cli getbestblockhash") false")

relayer_pid=$!
kill $relayer_pid
before_gap_height=$(docker exec -it bitcoind-node bitcoin-cli getblockchaininfo | jq ".headers")
echo $before_gap_height
# create 10 blocks new
docker exec -it bitcoind-node bitcoin-cli -generate 10


docker exec "$CONTAINER_ID" /bin/bash -c "sui client call --function latest_block_hash --module light_client --package '$PACKAGE_ID' --gas-budget 100000000 --args $LIGHT_CLIENT_ID --json"


# docker exec "$CONTAINER_ID" /bin/bash -c "sui client call --function insert_headers --module light_client --package '$PACKAGE_ID' --gas-budget 100000000 --args $LIGHT_CLIENT_ID '[0x${new_block_header}]' --json"

# after_gap_height=$(docker exec -it bitcoind-node bitcoin-cli getblockchaininfo | jq ".headers")

# echo $after_gap_height
echo "E2E Gap handler tests completed."
