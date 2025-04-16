#!/bin/bash
set -e
CONTAINER_ID="sui-node"

echo "Running E2E tests..."

echo "Start relayer to sync first time"

./e2e/sync_and_validate_status.sh

echo "Finished first sync and stop relayer"

echo "==============================="

echo "Other party update light client"
echo "Insert one header to lc"
docker exec -it bitcoind-node bitcoin-cli -generate 1 > /dev/null 2>&1
sleep 2
new_block_header=$(docker exec bitcoind-node  /bin/bash -c "bitcoin-cli getblockheader $(docker exec bitcoind-node  /bin/bash -c "bitcoin-cli getbestblockhash") false")
docker exec "$CONTAINER_ID" /bin/bash -c "sui client call --function insert_headers --module light_client --package '$PACKAGE_ID' --gas-budget 100000000 --args $LIGHT_CLIENT_ID '[0x${new_block_header}]' --json" > /dev/null 2>&1
sleep 10
echo "Inserted header"

echo "==============================="
echo "Create more 10 block"
docker exec -it bitcoind-node bitcoin-cli -generate 10 > /dev/null 2>&1
sleep 2
echo "=============================="
echo "Start relayer to sync second time"
./e2e/sync_and_validate_status.sh
echo "Finished second sync and stop relayer"
echo "Everything work well"

echo "E2E Gap handler tests completed."

exit 0
