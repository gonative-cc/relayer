#!/bin/bash

datadir="./regtest_demo"
walletsdir="wallets"

bitcoin-cli -datadir=$datadir stop

rm -rf $datadir

mkdir -p $datadir

conf_file="$datadir/bitcoin.conf"
cat << EOF > $conf_file
regtest=1
server=1
daemon=1
rpcuser=testuser
rpcpassword=testpassword
deprecatedrpc=warnings
[regtest]
rpcport=18443
port=18444
EOF

bitcoind -daemon -conf=./bitcoin.conf -datadir=$datadir
sleep 5

if [ ! -d "$walletsdir" ]; then
  mkdir -p "$walletsdir"
  for i in $(seq 1 5); do
    wallet_name="wallet${i}"
    backup_file="./wallets/$wallet_name.dat"
    bitcoin-cli -regtest -datadir=$datadir createwallet "$wallet_name"
    bitcoin-cli -regtest -datadir=$datadir -rpcwallet="$wallet_name" backupwallet "$backup_file"
    echo "Created and backed up wallet: $wallet_name"
  done
fi

for i in $(seq 1 5); do
  wallet_name="wallet${i}"

  bitcoin-cli -datadir=$datadir restorewallet "$wallet_name" "./wallets/$wallet_name.dat" > /dev/null
  bitcoin-cli -regtest -datadir=$datadir generatetoaddress 101 "$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet="$wallet_name" getnewaddress)" > /dev/null

  recipient_address=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet="$wallet_name" getnewaddress)
  utxo_txid=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet="$wallet_name" listunspent | jq -r '.[0].txid')
  utxo_vout=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet="$wallet_name" listunspent | jq -r '.[0].vout')

  rawtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet="$wallet_name" createrawtransaction '[{"txid":"'"$utxo_txid"'","vout":'"$utxo_vout"'}]' '{"'"$recipient_address"'":0.025}')
  signedtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet="$wallet_name" signrawtransactionwithwallet "$rawtx")
  final_signed_tx=$(echo "$signedtx" | jq -r '.hex')

  echo "$wallet_name:"
  echo "Raw Transaction: $rawtx"
  echo "Signed Transaction: $final_signed_tx"
  echo 
done