# Run bitcoind regtest node 

## Create bitcoind node 
Create a new bitcoind node with snapshot data 

```bash
make create-bitcoind
```

## Stop bitcoind node 

```bash
make stop-bitcoind
```

## Start bitcoind node 
You can start bitcoind node again. This command will keep any state you change on bitcoind node.

## Restart bitcoind node 
Restart bitcoind will create node with snapshot data.
```bash
make restart-bitcoind
```

## Interact with bitcoind node 

```bash
 docker exec -it bitcoind-node bitcoin-cli -regtest -rpcport=8332 <args>
```


