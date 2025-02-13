# Bitcoin SPV relayer

The code is adapted from [https://github.com/babylonchain/vigilante/tree/dev/reporter](https://github.com/babylonchain/).

This relayer is responsible for:

- syncing the latest BTC blocks with a BTC node
- detecting and reporting inconsistency between BTC blockchain and Lightclient header chain

## Usage

1. Setup the docker containers for `bitcoind` node and `bitcoin-lightclient`, refer to `../contrib/bitcoin-mock.md` for instructions.

2. Build the `bitcoin-spv` binary:

    ```bash
    go build ./cmd/bitcoin-spv
    ```

3. Start the relayer:

    ```bash
    ./bitcoin-spv start --config ./sample-bitcoin-spv.yml
    ```
