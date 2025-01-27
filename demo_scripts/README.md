# Regtest Demo Script

This script sets up a Bitcoin regtest environment, creates wallets, funds them, and generates raw and signed transactions. The wallets are being saved in `/wallets` directory, so the transactions and the signatures are repetable across many different runs.

## Prerequisites

- `bitciond`
- `jq`

## How to Run

1. Make sure the scirpt `start_local_network.sh` is executable
2. Execute the script with `./start_local_network.sh`.

## What the Script Does

1. Sets up regtest environment:
    - Stops any existing `bitcoind` process.
    - Removes the old `regtest_demo` data directory
    - Creates a new `regtest_demo` directory.
    - Creates a `bitcoin.conf` file for regtest with predefined settings.
    - Starts `bitcoind` in regtest mode.
2. Creates wallets:
    - Creates a `wallets` directory if it doesn't exist.
    - Creates 5 new wallets
    - Backs up each wallet  in the `wallets` directory.
3. Funds wallets and generates transactions:
    - Restores each wallet from its backup.
    - Generates 101 blocks to fund each wallet.
    - Creates a raw transaction sending 0.025 BTC to a new address in the same wallet.
    - Signs the raw transaction.
    - utputs the raw and signed transactions for each wallet.
