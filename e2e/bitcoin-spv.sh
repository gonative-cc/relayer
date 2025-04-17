#!/usr/bin/env bash
set -e # Exit immediately when the script failed
echo "Starting bitcoind node..."
make bitcoind-init
cd contrib/
docker compose up -d
cd ../

echo "Starting bitcoin-spv relayer..."
go build ./cmd/bitcoin-spv/
./bitcoin-spv start --config ./e2e-bitcoin-spv.yml 2>stderr.log &
RELAYER_PID=$! # gets pid of last background process
echo "Started bitcoin-spv relayer with PID $RELAYER_PID"

echo "Waiting for bitcoin-spv relayer to bootstrap..."
sleep 10

# make sure bitcoind node is up and running
if ! docker exec -it bitcoind-node bitcoin-cli -regtest getblockchaininfo; then
  echo "ERROR: failed to start bitcoind node"
  exit 1
fi

# query the lightclient for the latest block
chaintip_response_before=$(sui client call --function latest_block_hash --module light_client --package $PACKAGE_ID --gas-budget 10000000 --args $OBJECT_ID --dev-inspect --json)

chaintip_height_before=$(echo "$chaintip_response_before" | jq -r '.events[].parsedJson.height')

# produce a block
docker exec -it bitcoind-node bitcoin-cli -generate 1

# query the lightclient again for the latest block
chaintip_response_after=$(sui client call --function latest_block_hash --module light_client --package $PACKAGE_ID --gas-budget 10000000 --args $OBJECT_ID --dev-inspect --json)

# parse the JSON to extract values using jq
chaintip_height_after=$(echo "$chaintip_response_after" | jq -r '.events[].parsedJson.height')

# assert that the lightclient block has increased by 1
if [[ $((chaintip_height_after-chaintip_height_before)) != 1 ]]; then
  echo "ERROR: light client didn't update correctly: the latest confirmed block didn't change"
  echo "Before Hash: $chaintip_hash_before"
  echo "Before Height: $chaintip_height_before"
  echo "After Hash: $chaintip_hash_after"
  echo "After Height: $chaintip_height_after"
fi

# read the log file of relayer
if [ ! -s "stderr.log" ]; then
  echo "ERROR $(cat stderr.log)"
else
  echo "SUCCESS: bitcoin block relayed successfully"
fi

rm -f stderr.log

kill $RELAYER_PID
