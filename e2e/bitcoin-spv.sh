#!/usr/bin/env bash

# spin up bitcoind node
echo "Starting bitcoind node and bitcoin-lightclient..."
make bitcoind-init
cd contrib/
docker compose up -d
cd ../
echo "Started bitcoind node and bitcoin-lightclient"

# spin up bitcoin-spv relayer
echo "Starting bitcoin-spv relayer..."
go build ./cmd/bitcoin-spv/
./bitcoin-spv bitcoin-spv --config ./sample-bitcoin-relayer.yml

# make sure bitcoind node is up and running
docker exec -it bitcoind-node bitcoin-cli -regtest getblockchaininfo

# query the lightclient for the latest block
chaintip_response_before=$(curl --user user:password --data-binary '{"jsonrpc":"1.0","id":"curltest","method":"get_header_chain_tip","params":[]}' -H 'content-type: text/plain;' http://127.0.0.1:9797)

# parse the JSON to extract values using jq
chaintip_hash_before=$(echo "$chaintip_response_before" | jq -r '.result.Hash')
chaintip_height_before=$(($(echo "$chaintip_response_before" | jq -r '.result.Height')))

# produce a block
docker exec -it bitcoind-node bitcoin-cli -generate 1

# query the lightclient again for the latest block
chaintip_response_after=$(curl --user user:password --data-binary '{"jsonrpc":"1.0","id":"curltest","method":"get_header_chain_tip","params":[]}' -H 'content-type: text/plain;' http://127.0.0.1:9797)

# parse the JSON to extract values using jq
chaintip_hash_after=$(echo "$chaintip_response_after" | jq -r '.result.Hash')
chaintip_height_after=$(($(echo "$chaintip_response_after" | jq -r '.result.Height')))

# assert that the lightclient block has increased by 1
echo "Before Hash: $chaintip_hash_before"
echo "Before Height: $chaintip_height_before"

echo "After Hash: $chaintip_hash_after"
echo "After Height: $chaintip_height_after"

if (( chaintip_height_after - chaintip_height_before == 1 )); then
  echo "The values have a difference of 1."
else
  echo "The values do not have a difference of 1."
fi

# check if the spv proof has been submitted and is accepted

