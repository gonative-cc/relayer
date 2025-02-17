# Bitcoin-SPV configuration

## Sample configuration file

```yaml
relayer:
  log-format: "auto" # (json|auto|console|logfmt)
  log-level: "debug" # (debug|warn|error|panic|fatal)
  retry-sleep-duration: 5s
  max-retry-sleep-duration: 5m
  netparams: simnet
  cache-size: 1000
  headers-chunk-size: 100
btc:
  no-client-tls: true
  ca-file: $HOME/.btcd/rpc.cert
  endpoint: localhost:18443
  net-params: regtest
  username: user
  password: password
  btc-backend: bitcoind # {btcd, bitcoind}
  zmq-seq-endpoint: tcp://127.0.0.1:28331
native:
  rpc-endpoint: http://localhost:9797
```

## Relayer config

- `log-format`: Format for log output (json|auto|console|logfmt)
- `log-level`: Logging level (debug|warn|error|panic|fatal)
- `retry-sleep-duration`: Duration to wait between retry attempts
- `max-retry-sleep-duration`: Maximum duration to wait between retry attempts
- `netparams`: Bitcoin network parameters (mainnet|testnet|simnet|regtest)
- `cache-size`: Size of the block headers cache
- `headers-chunk-size`: Number of headers to request in a single chunk

## Bitcoin node config

- `no-client-tls`: Disable TLS for client connections to Bitcoin node
- `ca-file`: Path to Bitcoin node's TLS certificate file
- `endpoint`: Bitcoin node RPC endpoint address
- `net-params`: Bitcoin network parameters (mainnet|testnet|simnet|regtest)
- `username`: RPC username for Bitcoin node authentication
- `password`: RPC password for Bitcoin node authentication
- `btc-backend`: Bitcoin node implementation to use (btcd|bitcoind)
- `zmq-seq-endpoint`: ZeroMQ sequence notification endpoint for Bitcoin node

## Bitcoin-lightclient config

- `rpc-endpoint`: RPC endpoint address for the Bitcoin light client
