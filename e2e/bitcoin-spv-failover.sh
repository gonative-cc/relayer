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
chaintip_response_before=$(curl --user user:password --data-binary '{"jsonrpc":"1.0","id":"curltest","method":"get_header_chain_tip","params":[]}' -H 'content-type: text/plain;' http://127.0.0.1:9797)

chaintip_hash_before=$(echo "$chaintip_response_before" | jq -r '.result.Hash')
chaintip_height_before=$(echo "$chaintip_response_before" | jq -r '.result.Height')

# stop the relayer
echo "Stopping bitcoin-spv relayer..."
kill $RELAYER_PID_OLD

# produce 100 blocks
docker exec -it bitcoind-node bitcoin-cli -generate 100

# query the lightclient again for the latest block
chaintip_response_after_stop=$(curl --user user:password --data-binary '{"jsonrpc":"1.0","id":"curltest","method":"get_header_chain_tip","params":[]}' -H 'content-type: text/plain;' http://127.0.0.1:9797)

# parse the JSON to extract values using jq
chaintip_hash_after_stop=$(echo "$chaintip_response_after_stop" | jq -r '.result.Hash')
chaintip_height_after_stop=$(echo "$chaintip_response_after_stop" | jq -r '.result.Height')

# start the relayer again
echo "Starting bitcoin-spv relayer again..."
go build ./cmd/bitcoin-spv/
./bitcoin-spv start --config ./sample-bitcoin-relayer.yml 2>stderr.log &
RELAYER_PID_NEW=$! # gets pid of last background process
echo "Started bitcoin-spv relayer again with PID $RELAYER_PID_NEW"

echo "Waiting for bitcoin-spv relayer to bootstrap..."
sleep 10

# query the lightclient again for the latest block
chaintip_response_after_restart=$(curl --user user:password --data-binary '{"jsonrpc":"1.0","id":"curltest","method":"get_header_chain_tip","params":[]}' -H 'content-type: text/plain;' http://127.0.0.1:9797)

# parse the JSON to extract values using jq
chaintip_hash_after_restart=$(echo "$chaintip_response_after_restart" | jq -r '.result.Hash')
chaintip_height_after_restart=$(echo "$chaintip_response_after_restart" | jq -r '.result.Height')

# assert that the lightclient block has increased by exactly 100
if [[ $((chaintip_height_after-chaintip_height_before)) != 100 ]]; then
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

kill $RELAYER_PID_NEW
