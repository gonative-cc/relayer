# Bitcoin-SPV

This relayer is responsible for syncing the SPV Light Client with a BTC node.

- [Configuration](./configuration.md)

## Setup

### Dependencies

- Docker, or [Docker Desktop](https://www.docker.com/products/docker-desktop) for an easy UX
- [Go language](https://golang.org/dl/)

### Run bitcoind node and bitcoin-lightclient

- Setup the docker containers for `bitcoind` node and `bitcoin-lightclient`, refer to [../contrib/bitcoin-mock.md](../contrib/bitcoin-mock.md) for instructions.

## Usage

### Make sure bitcoind is working

Query blockchain info:

```bash
bitcoin-cli -regtest -rpcuser=user -rpcpassword=password getblockchaininfo
```

### Start the relayer

- Build the `bitcoin-spv` binary:

    ```bash
    go build ./cmd/bitcoin-spv
    ```

- Start the relayer:

    ```bash
    ./bitcoin-spv start --config ./sample-bitcoin-spv.yml
    ```

## Relayer Flow

Following diagram explains how the bitcoin-SPV relayer interacts with `BitcoinNode` and `LightClient` and how data flows from Bitcoin node to Light Client through the SPV relayer.

### Connecting

```mermaid
flowchart TD
    A["Bitcoin full node"] <-- 1 Connect --> B("Native SPV relayer")
    B <-- 2 Connect --> D["SPV lightclient"]
    B -- 3 Sync blocks --> D
    B <-. 4 New events listen .-> A
```

### Sending block headers

```mermaid
flowchart TD
    A["Bitcoin full node"] -. 1 New blocks .-> B("Native SPV relayer")
    B -- 2 Send blockheader --> D["SPV lightclient"]
```
