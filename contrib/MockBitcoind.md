# Run bitcoind regtest node

## Requirements

Make sure you have docker and docker-compose-plugin installed and docker service is running.

## Initialization

```sh
make bitcoind-init
```

Start a new terminal where you will run a container with bitcoind, by default the nativewallet will be loaded:

```sh
cd contrib; docker compose up
```

## Stop bitcoind node

To stop the node just press `ctrl-c` in the terminal running a node.

To remove a container:

```sh
docker compose down
```

## Interact with the bitcoind node

```sh
docker exec -it bitcoind-node bitcoin-cli -regtest <args>
```

Or enter bash in the container:

```sh
docker exec -it bitcoind-node /bin/bash
```

Then you can generate a block:

```sh
bitcoin-cli -regtest generate
```

If RPC params are required, you can provide them:

```sh
bitcoin-cli -regtest -rpcuser=user -rpcpassword=password generate
```

## Reference

- [Running Bitcoind with ZMQ](https://bitcoindev.network/accessing-bitcoins-zeromq-interface/)
- [Bitlights Labs dev env](https://blog.bitlightlabs.com/posts/setup-local-development-env-regtest)
