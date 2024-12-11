# Run btcd simnet mode

## Install btcd and btcwallet

Follow the instruction in [btcd README.md](https://github.com/btcsuite/btcd?tab=readme-ov-file#installation)
and [btcwallet README](https://github.com/btcsuite/btcwallet?tab=readme-ov-file#installation-and-updating) for install btcd and btcwallet

## Run btcd with simnet mode

### Bootstrap the wallet and btc node

First you need start the btcd in simnet mode

```bash
btcd --simnet --rpcuser=user --rpcpass=password
```

Create btc wallet. Please remember password we can use this somewhere.

```bash
btcwallet --simnet --create
```

Connect you wallet to btcd simmode. The password and user here must be the btcd password/user pair.
Open a new terminal and run command below:

```bash
btcwallet --simnet -u=user -P=password
```

Create a new address we will use this for `--miningaddr`, the address receive btc when mining a new block.
Open a new terminal and run command below:

```bash
btcctl --simnet --wallet --rpcuser=user --rpcpass=password getnewaddress
```

Copy the address created by command above. Shutdown the btcd service and run with command below:

```bash
btcd --simnet --rpcuser=user --rpcpass=password --miningaddr=<address>
```

Right now, everytime we mine a new block, the minner address should receive some bitcoin.
We can use this for testing and development.

## Generate a new block

Generate 100 blocks

```bash
btcctl --simnet --wallet --rpcuser=user --rpcpass=password generate 100
```

## Check information

Blockchain information

```bash
btcctl --simnet --wallet --rpcuser=user --rpcpass=password getblockchaininfo
```

Miner balance

```bash
btcctl --simnet --wallet --rpcuser=user --rpcpass=password getbalance
```
