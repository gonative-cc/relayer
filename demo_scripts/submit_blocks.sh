#!/bin/bash

START_HEIGHT=0
END_HEIGHT=201 # Include this

PACKAGE_ID="0x267e10950a7242bb2958833ec9a1af49c50c7dd795794d9f903db61b14be70d6"
STATE_ID="0xae30240c53cca6ebab022bd996bd2a8e44490db6c71e393564ae241f0585529c"
MODULE_NAME="bitcoin_executor"
FUNCTION="execute_block"
GAS_BUDGET="100000000"

echo "Submitting blocks for execution from height $START_HEIGHT to $END_HEIGHT..."

for (( HEIGHT=START_HEIGHT; HEIGHT<=END_HEIGHT; HEIGHT++ )); do
    echo "Processing block at height: $HEIGHT"

    BLOCK_HASH=$(bitcoin-cli -regtest getblockhash $HEIGHT 2>/dev/null)

    if [ -z "$BLOCK_HASH" ]; then
        echo "Error: Could not get block hash for height $HEIGHT"
        continue
    fi

    BLOCK_HEX=$(bitcoin-cli -regtest getblock "$BLOCK_HASH" 0 2>/dev/null)

    if [ -z "$BLOCK_HEX" ]; then
        echo "Error: Could not get block hex for hash $BLOCK_HASH (height $HEIGHT)"
        continue
    fi

    sui client call --package "$PACKAGE_ID" --module "$MODULE_NAME" --function "$FUNCTION" --args "$STATE_ID" 0x"${BLOCK_HEX}" --gas-budget "$GAS_BUDGET"
done

echo "--------------------------------------------------------------------------"
echo "Script finished submiting blocks for execution for blocks $START_HEIGHT to $END_HEIGHT"
