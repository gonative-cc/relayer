#!/bin/bash
START_HEIGHT=0

echo "Starting Block Height: $START_HEIGHT"

INIT_HEADERS='['
for (( i=$START_HEIGHT; i<=$START_HEIGHT+11; i++ )); do
  BLOCK_HASH=$(bitcoin-cli getblockhash $i)
  HEADER_HEX=$(bitcoin-cli getblockheader $BLOCK_HASH false)
  INIT_HEADERS+="0x"$HEADER_HEX
  if (( i < $START_HEIGHT+11 )); then
     INIT_HEADERS+=","
  fi
done
INIT_HEADERS+=']'
echo "INIT_HEADERS: $INIT_HEADERS"