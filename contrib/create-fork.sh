#!/bin/bash


function createFork() {
    forkName=$1
    forkLength=$2
    echo $forkName
    cp -rf ./bitcoind-snapshot/ $forkName
    BITCOIND_DATA=$forkName docker-compose up -d
    docker exec -it bitcoind-node bitcoin-cli -generate $forkLength
}

function extractFork() {
    start=$1
    end=$2
    for ((i=$start; i<=$end; i++)); do
	myhash=$(docker exec -it bitcoind-node bitcoin-cli getblockhash $i)
	echo $myhash
	ver='0'
	tmp=$(echo $myhash)
	echo $tmp
	echo "$myhash $ver"
	echo $h
	# docker exec -it bitcoind-node bitcoin-cli getblockheader $h false
    done
}


case "$1" in
  create)
      createFork $2 $3
    ;;
  extract)
      extractFork $2 $3
    ;;
  *)
    echo "Invalid option."
    exit 1
    ;;
esac
