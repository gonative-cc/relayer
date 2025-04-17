#!/bin/bash
set -e
CONTAINER_ID="sui-node"
E2E_YAML_CONFIG=./e2e-bitcoin-spv.yml
LIGHT_CLIENT_ID="$(yq -r '.sui.lc_object_id' $E2E_YAML_CONFIG)"
PACKAGE_ID="$(yq -r '.sui.lc_package_id' $E2E_YAML_CONFIG)"


echo "Running E2E tests..."

echo "Start relayer to sync first time"

docker exec -i bitcoind-node bitcoin-cli -generate 100 > /dev/null 2>&1
sleep 5
./e2e/sync_and_validate_status.sh

echo "Finished first sync and stopped the relayer"

echo "==============================="

echo "Other party updates light client"
echo "Attempt Inserting one header to lc"
docker exec -i bitcoind-node bitcoin-cli -generate 1 > /dev/null 2>&1
sleep 5
new_block_header=$(docker exec bitcoind-node  /bin/bash -c "bitcoin-cli getblockheader $(docker exec bitcoind-node  /bin/bash -c "bitcoin-cli getbestblockhash") false")
docker exec "$CONTAINER_ID" /bin/bash -c "sui client call --function insert_headers --module light_client --package '$PACKAGE_ID' --gas-budget 100000000 --args $LIGHT_CLIENT_ID '[0x${new_block_header}]' --json" > /dev/null 2>&1
sleep 10
echo "Inserted header"

echo "==============================="
echo "Generate 10 more blocks"
docker exec -i bitcoind-node bitcoin-cli -generate 10 > /dev/null 2>&1
sleep 2
echo "=============================="
echo "Start relayer to sync second time"
./e2e/sync_and_validate_status.sh
echo "Finished second sync and stopped relayer"
echo "Everything worked well"

echo "E2E Gap handler tests completed."

exit 0
