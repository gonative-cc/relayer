#!/bin/bash

# Create fork
function createFork() {
    forkName="./fork-"$1
    forkLength=$2

    cp -rf ./bitcoind-snapshot/ $forkName
    BITCOIND_DATA=$forkName docker-compose up -d
    docker exec -it bitcoind-node bitcoin-cli -generate $forkLength
    docker-compose down
}

# Extract block between range 
function extractFork() {
    forkName="./fork-"$1
    start=$2
    end=$3

    BITCOIND_DATA=$forkName docker-compose up -d
    docker exec -it bitcoind-node bitcoin-cli -generate $forkLength

    for ((i=$start; i<=$end; i++)); do
	hash=$(docker exec -it bitcoind-node bitcoin-cli getblockhash $i | tr -d "\r\n")
	docker exec -it bitcoind-node bitcoin-cli getblockheader $hash false
    done
    docker-compose down
}


case "$1" in
  create)
      createFork $2 $3
    ;;
  extract)
      extractFork $2 $3 $4
    ;;
  *)
    echo "Invalid option."
    exit 1
    ;;
esac
