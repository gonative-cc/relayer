#!/usr/bin/env bash

echo "Starting bitcoind node and bitcoin-lightclient..."
make bitcoind-init
cd contrib/
docker compose up -d
cd ../
echo "Started bitcoind node and bitcoin-lightclient"

echo "Starting bitcoin-spv relayer..."
go build ./cmd/bitcoin-spv/
./bitcoin-spv start --config ./sample-bitcoin-relayer.yml 2>stderr.log &
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
chaintip_response_before=$(curl --user user:password --data-binary '{"jsonrpc":"1.0","id":"curltest","method":"get_header_chain_tip","params":[]}' -H 'content-type: text/plain;' http://127.0.0.1:9797)

chaintip_hash_before=$(echo "$chaintip_response_before" | jq -r '.result.Hash')
chaintip_height_before=$(echo "$chaintip_response_before" | jq -r '.result.Height')

# produce a block
docker exec -it bitcoind-node bitcoin-cli -generate 1

# query the lightclient again for the latest block
chaintip_response_after=$(curl --user user:password --data-binary '{"jsonrpc":"1.0","id":"curltest","method":"get_header_chain_tip","params":[]}' -H 'content-type: text/plain;' http://127.0.0.1:9797)

# parse the JSON to extract values using jq
chaintip_hash_after=$(echo "$chaintip_response_after" | jq -r '.result.Hash')
chaintip_height_after=$(echo "$chaintip_response_after" | jq -r '.result.Height')

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
