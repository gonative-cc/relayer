# Run bitcoind regtest node

## Create bitcoind node

Create a new bitcoind node with snapshot data

```bash
make bitcoind-run
```

## Stop bitcoind node

```bash
make bitcoind-stop
```

## Start bitcoind node

You can start bitcoind node again. This command will keep any state changes on the bitcoind node.

```bash
make bitcoind-start
```

## Restart bitcoind node

Restarting bitcoind will create node with a snapshot data.

```bash
make bitcoind-reinit
```

## Interact with bitcoind node

```bash
 docker exec -it bitcoind-node bitcoin-cli -regtest -rpcport=8332 <args>
```
