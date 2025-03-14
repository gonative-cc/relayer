#!/usr/bin/env bash
set -e # Exit immediately when the script failed
echo "Starting bitcoind node and bitcoin-lightclient..."
make bitcoind-init
cd contrib/
docker compose up -d
cd ../

echo "Starting bitcoin-spv relayer..."
go build ./cmd/bitcoin-spv/
./bitcoin-spv start --config ./sample-bitcoin-relayer.yml 2>stderr.log &
RELAYER_PID_OLD=$! # gets pid of last background process
echo "Started bitcoin-spv relayer with PID $RELAYER_PID_OLD"

echo "Waiting for bitcoin-spv relayer to bootstrap..."
sleep 10

# make sure bitcoind node is up and running
if ! docker exec -it bitcoind-node bitcoin-cli -regtest getblockchaininfo; then
  echo "ERROR: failed to start bitcoind node"
  exit 1
fi

# query the lightclient for the latest block
chaintip_response_before=$(sui client call --function latest_block --module light_client --package $PACKAGE_ID --gas-budget 10000000 --args $OBJECT_ID --dev-inspect --json)

chaintip_height_before=$(echo "$chaintip_response_before" | jq -r '.events[].parsedJson.height')

# stop the relayer
echo "Stopping bitcoin-spv relayer..."
kill $RELAYER_PID_OLD

# produce 100 blocks
docker exec -it bitcoind-node bitcoin-cli -generate 100

# query the lightclient again for the latest block
chaintip_response_after_stop=$(sui client call --function latest_block --module light_client --package $PACKAGE_ID --gas-budget 10000000 --args $OBJECT_ID --dev-inspect --json)

# parse the JSON to extract values using jq
chaintip_height_after_stop=$(echo "$chaintip_response_after_stop" | jq -r '.events[].parsedJson.height')

# start the relayer again
echo "Starting bitcoin-spv relayer again..."
go build ./cmd/bitcoin-spv/
./bitcoin-spv start --config ./sample-bitcoin-relayer.yml 2>stderr.log &
RELAYER_PID_NEW=$! # gets pid of last background process
echo "Started bitcoin-spv relayer again with PID $RELAYER_PID_NEW"

echo "Waiting for bitcoin-spv relayer to bootstrap..."
sleep 50 # NOTE: long wait cause waiting for relayer to submit all 100 block headers

# query the lightclient again for the latest block
chaintip_response_after_restart=$(sui client call --function latest_block --module light_client --package $PACKAGE_ID --gas-budget 10000000 --args $OBJECT_ID --dev-inspect --json)

# parse the JSON to extract values using jq
chaintip_height_after_restart=$(echo "$chaintip_response_after_restart" | jq -r '.events[].parsedJson.height')

# assert that the lightclient block has increased by exactly 100
if [[ $((chaintip_height_after_restart-chaintip_height_before)) != 100 ]]; then
  echo "ERROR: light client didn't update correctly: the latest confirmed block didn't change"
  echo "Before Height: $chaintip_height_before"
  echo "After Height: $chaintip_height_after_restart"
fi

# stop the relayer
echo "Stopping bitcoin-spv relayer..."
kill $RELAYER_PID_NEW

# testing restart relayer after few more blocks have been posted

# manually send the initial 20 block headers using sui client
for i in {1..20}; do
  # generate a block
  docker exec -it bitcoind-node bitcoin-cli -generate 1

  # query latest blockhash
  latest_blockhash=$(docker exec -it bitcoind-node bitcoin-cli -regtest getbestblockhash)

  # query the blockheader of that blockhash
  latest_blockheader=$(docker exec -it bitcoind-node bitcoin-cli -regtest getblockheader $latest_blockhash false)

  # send the blockheader to sui LC using sui client
  sui client call --function insert_headers --module light_client --package $PACKAGE_ID --gas-budget 100000000 --args $OBJECT_ID ['0x$latest_blockheader']
done

# produce 80 blocks
docker exec -it bitcoind-node bitcoin-cli -generate 80

# start relayer
echo "Starting bitcoin-spv relayer..."
go build ./cmd/bitcoin-spv/
./bitcoin-spv start --config ./sample-bitcoin-relayer.yml 2>stderr.log &
RELAYER_PID_AGAIN=$! # gets pid of last background process
echo "Started bitcoin-spv relayer with PID $RELAYER_PID_AGAIN"

echo "Waiting for second bitcoin-spv relayer to bootstrap..."
sleep 50 # NOTE: long wait cause waiting for relayer to submit all 80 block headers

# query the lightclient again for the latest block
chaintip_response_after_concurrent=$(sui client call --function latest_block --module light_client --package $PACKAGE_ID --gas-budget 10000000 --args $OBJECT_ID --dev-inspect --json)

# parse the JSON to extract values using jq
chaintip_height_after_concurrent=$(echo "$chaintip_response_after_concurrent" | jq -r '.events[].parsedJson.height')

# assert that the lightclient block has increased by exactly 200
if [[ $((chaintip_height_after_concurrent-chaintip_height_before)) != 200 ]]; then
  echo "ERROR: light client didn't update correctly: the latest confirmed block didn't change"
  echo "Before Height: $chaintip_height_before"
  echo "After Height: $chaintip_height_after_concurrent"
fi

# read the log file of relayer
if [ ! -s "stderr.log" ]; then
  echo "ERROR $(cat stderr.log)"
else
  echo "SUCCESS: bitcoin block relayed successfully"
fi

rm -f stderr.log

# stop the relayer
echo "Stopping bitcoin-spv relayer..."
kill $RELAYER_PID_AGAIN
