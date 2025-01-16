# Run bitcoind regtest node

## Requirements

- Docker Engine >= 27.0
- docker-compose-plugin (don't use the legacy python docker-compose)
- docker service running; [install instructions](https://docs.docker.com/engine/install/).

There are two leading Bitcoin implementation, each comes with a mode to run simulations that helps to use it
as a mock, with predefined snapshot and data:

- [Bitcoind](https://github.com/bitcoin/bitcoin) with the `-regtest` mode
- [Btcd](https://github.com/btcsuite/btcd) with the `-simtest` mode

## Bitcoind

Bitcoind is a binary from the reference implementation

### Setting up

Firstly we need to copy the snapshot (to have a shared data):

```sh
make bitcoind-init
```

Start a new terminal where you will run a container with bitcoind, by default the nativewallet will be loaded:

```sh
cd contrib; docker compose up
```

To stop the node just press `ctrl-c` in the terminal running a node.

To remove a container:

```sh
docker compose down
```

### Interact with the bitcoind node

```sh
docker exec -it bitcoind-node bitcoin-cli -regtest <args>
```

Or enter bash in the container:

```sh
docker exec -it bitcoind-node /bin/bash
```

Then you can generate a block:

```sh
bitcoin-cli -regtest -generate <number-block>
```

If RPC params are required, you can provide them:

```sh
bitcoin-cli -regtest -rpcuser=user -rpcpassword=password -generate <number-block>
```

More information in [developer.bitcoin.org -> testing](https://developer.bitcoin.org/examples/testing.html).

### Create BTC fork for testing

In a few cases, we must create a BTC fork for testing. The create-fork.sh script helps you do this. We provide two functions:

#### Create fork

Creates a new fork starting at the latest block in the snapshot data.

```sh
./create-fork.sh create <fork-name> <number-block>
```

#### Extract fork

We can extract any block between a specific range. This command below returns the list of block headers in this range.

```sh
./create-fork.sh extract <fork-name> <start> <end>
```

### Reference

- [Running Bitcoind with ZMQ](https://bitcoindev.network/accessing-bitcoins-zeromq-interface/)
- [Bitlights Labs dev env](https://blog.bitlightlabs.com/posts/setup-local-development-env-regtest)

## Btcd

alternatively to Bitcoind we can use btcd.

### Install btcd and btcwallet

Follow the instruction in [btcd README.md](https://github.com/btcsuite/btcd?tab=readme-ov-file#installation)
and [btcwallet README](https://github.com/btcsuite/btcwallet?tab=readme-ov-file#installation-and-updating) to install btcd and btcwallet

### Bootstrap the wallet and btc node

First you need to start the btcd in simnet mode

```bash
btcd --simnet --rpcuser=user --rpcpass=password
```

Create btc wallet. Please remember the password.

```bash
btcwallet --simnet --create
```

Connect you wallet to btcd simmode. The password and user here must be the btcd password/user pair.
Open a new terminal and run:

```bash
btcwallet --simnet -u=user -P=password
```

Create a new address. We will use it as the `--miningaddr` parameter - an address that receives btc when mining a new block.
Open a new terminal and run:

```bash
btcctl --simnet --wallet --rpcuser=user --rpcpass=password getnewaddress
```

Copy the address created by the command above. Shutdown the btcd service and run:

```bash
btcd --simnet --rpcuser=user --rpcpass=password --miningaddr=<address>
```

Right now, everytime we mine a new block, the minner address should receive some bitcoin.
We can use this for testing and development.

### Generate a new block

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
