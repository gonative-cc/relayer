#!/bin/bash

datadir="./regtest_demo"
# backup_file="./wallets/wallet1.dat"

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

#Restore wallet
bitcoin-cli -datadir=$datadir restorewallet "wallet1" ./wallets/wallet1.dat

bitcoin-cli -regtest -datadir=$datadir generatetoaddress 101 $(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet1 getnewaddress)

recipient_address=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet1 getnewaddress)

utxo_txid=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet1 listunspent | jq -r '.[0].txid')
utxo_vout=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet1 listunspent | jq -r '.[0].vout')

rawtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet1 createrawtransaction '[{"txid":"'"$utxo_txid"'","vout":'"$utxo_vout"'}]' '{"'$recipient_address'":0.025}')
signedtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet1 signrawtransactionwithwallet "$rawtx")

final_signed_tx=$(echo "$signedtx" | jq -r '.hex')

echo "wallet1:"
echo "Raw Transaction: $rawtx"
echo "Signed Transaction: $final_signed_tx"
echo


#Restore wallet
bitcoin-cli -datadir=$datadir restorewallet "wallet2" ./wallets/wallet2.dat

bitcoin-cli -regtest -datadir=$datadir generatetoaddress 101 $(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet2 getnewaddress)

recipient_address=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet2 getnewaddress)

utxo_txid=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet2 listunspent | jq -r '.[0].txid')
utxo_vout=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet2 listunspent | jq -r '.[0].vout')

rawtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet2 createrawtransaction '[{"txid":"'"$utxo_txid"'","vout":'"$utxo_vout"'}]' '{"'$recipient_address'":0.025}')
signedtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet2 signrawtransactionwithwallet "$rawtx")

final_signed_tx=$(echo "$signedtx" | jq -r '.hex')

echo "wallet2:"
echo "Raw Transaction: $rawtx"
echo "Signed Transaction: $final_signed_tx"
echo 

#Restore wallet
bitcoin-cli -datadir=$datadir restorewallet "wallet3" ./wallets/wallet3.dat

bitcoin-cli -regtest -datadir=$datadir generatetoaddress 101 $(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet3 getnewaddress)

recipient_address=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet3 getnewaddress)

utxo_txid=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet3 listunspent | jq -r '.[0].txid')
utxo_vout=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet3 listunspent | jq -r '.[0].vout')

rawtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet3 createrawtransaction '[{"txid":"'"$utxo_txid"'","vout":'"$utxo_vout"'}]' '{"'$recipient_address'":0.025}')
signedtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet3 signrawtransactionwithwallet "$rawtx")

final_signed_tx=$(echo "$signedtx" | jq -r '.hex')

echo "wallet3:"
echo "Raw Transaction: $rawtx"
echo "Signed Transaction: $final_signed_tx"
echo 

#Restore wallet
bitcoin-cli -datadir=$datadir restorewallet "wallet4" ./wallets/wallet4.dat

bitcoin-cli -regtest -datadir=$datadir generatetoaddress 101 $(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet4 getnewaddress)

recipient_address=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet4 getnewaddress)

utxo_txid=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet4 listunspent | jq -r '.[0].txid')
utxo_vout=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet4 listunspent | jq -r '.[0].vout')

rawtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet4 createrawtransaction '[{"txid":"'"$utxo_txid"'","vout":'"$utxo_vout"'}]' '{"'$recipient_address'":0.025}')
signedtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet4 signrawtransactionwithwallet "$rawtx")

final_signed_tx=$(echo "$signedtx" | jq -r '.hex')

echo "wallet4:"
echo "Raw Transaction: $rawtx"
echo "Signed Transaction: $final_signed_tx"
echo 

#Restore wallet
bitcoin-cli -datadir=$datadir restorewallet "wallet5" ./wallets/wallet5.dat

bitcoin-cli -regtest -datadir=$datadir generatetoaddress 101 $(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet5 getnewaddress)

recipient_address=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet5 getnewaddress)

utxo_txid=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet5 listunspent | jq -r '.[0].txid')
utxo_vout=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet5 listunspent | jq -r '.[0].vout')

rawtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet5 createrawtransaction '[{"txid":"'"$utxo_txid"'","vout":'"$utxo_vout"'}]' '{"'$recipient_address'":0.025}')
signedtx=$(bitcoin-cli -regtest -datadir=$datadir -rpcwallet=wallet5 signrawtransactionwithwallet "$rawtx")

final_signed_tx=$(echo "$signedtx" | jq -r '.hex')

echo "wallet5:"
echo "Raw Transaction: $rawtx"
echo "Signed Transaction: $final_signed_tx"
echo 