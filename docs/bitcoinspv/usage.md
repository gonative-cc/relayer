# Bitcoin-SPV usage

## Make sure bitcoind is working

- Query blockchain info:

    ```bash
    bitcoin-cli -regtest -rpcuser=user -rpcpassword=password getblockchaininfo
    ```

## Make sure bitcoin-lightclient is working

- Ping lightclient RPC:

    ```bash
    curl --user user:password --data-binary '{"jsonrpc":"1.0","id":"curltest","method":"ping","params":[]}' -H 'content-type: text/plain;' http://127.0.0.1:9797
    ```

## Start the relayer

- Build the `bitcoin-spv` binary:

    ```bash
    go build ./cmd/bitcoin-spv
    ```

- Start the relayer:

    ```bash
    ./bitcoin-spv start --config ./sample-bitcoin-spv.yml
    ```
