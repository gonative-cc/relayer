# Bitcoin-SPV configuration

## Sample configuration file

```yaml
relayer:
  log-format: "auto" # (json|auto|console|logfmt)
  log-level: "debug" # (debug|warn|error|panic|fatal)
  retry-sleep-duration: 5s # Duration to wait between retry attempts
  max-retry-sleep-duration: 5m # Maximum duration to wait between retry attempts
  netparams: regtest # (mainnet|testnet|simnet|regtest)
  cache-size: 1000 # Size of the block headers cache
  headers-chunk-size: 100 # Number of headers posted to lightclient in a single chunk
  process-block-timeout: 20 # Timeout duration for processing a single block, after which the context will be canceled
btc:
  no-client-tls: true # Disable TLS for client connections to Bitcoin node
  ca-file: $HOME/.btcd/rpc.cert # Path to Bitcoin node's TLS certificate file
  endpoint: localhost:18443 # Bitcoin node RPC endpoint address
  net-params: regtest # (mainnet|testnet|simnet|regtest)
  username: user # RPC username for Bitcoin node
  password: password # RPC password for Bitcoin node
  btc-backend: bitcoind # {btcd, bitcoind}
  zmq-seq-endpoint: tcp://127.0.0.1:28331 # ZeroMQ sequence notification endpoint for Bitcoin node
native:
  rpc-endpoint: http://localhost:9797 # RPC endpoint address for the Bitcoin light client
```
