#!/bin/bash
START_HEIGHT=2982713

echo "Starting Block Height: $START_HEIGHT"

INIT_HEADERS='['
for (( i=$START_HEIGHT; i<=$START_HEIGHT+10; i++ )); do
  BLOCK_HASH=$(bitcoin-cli -testnet getblockhash $i)
  HEADER_HEX=$(bitcoin-cli -testnet getblockheader $BLOCK_HASH false)
  INIT_HEADERS+="0x"$HEADER_HEX
  if (( i < $START_HEIGHT+9 )); then
     INIT_HEADERS+=","
  fi
done
INIT_HEADERS+=']'
echo "INIT_HEADERS: $INIT_HEADERS"