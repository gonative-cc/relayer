#!/bin/bash

datadir="./regtest_demo"
backup_file="./mydemowallet.dat"

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
[regtest]
rpcport=18443
port=18444
EOF

bitcoind -daemon -conf=./bitcoin.conf -datadir=$datadir
sleep 5

#Restore wallet
bitcoin-cli -datadir=$datadir restorewallet "mydemowallet" "$backup_file"

bitcoin-cli -regtest -datadir=$datadir generatetoaddress 101 $(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=mydemowallet getnewaddress)

recipient_address=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=mydemowallet getnewaddress)

utxo_txid=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=mydemowallet listunspent | jq -r '.[0].txid')
utxo_vout=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=mydemowallet listunspent | jq -r '.[0].vout')

rawtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=mydemowallet createrawtransaction '[{"txid":"'"$utxo_txid"'","vout":'"$utxo_vout"'}]' '{"'$recipient_address'":0.025}')
signedtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=mydemowallet signrawtransactionwithwallet "$rawtx")

final_signed_tx=$(echo "$signedtx" | jq -r '.hex')

echo "Raw Transaction: $rawtx"
echo "Signed Transaction: $final_signed_tx"