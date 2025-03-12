# Sui E2E Testing Scripts

This document describes the purpose of each script used in the end-to-end testing workflow.  These scripts are designed to be used with a Docker Compose environment defined in `contrib/docker-compose.yaml` and a pre-configured Sui network snapshot in `contrib/sui-snapshot`.

## Scripts

### `e2e/start_services.sh`

Starts the Docker Compose services (`sui-node` and `bitcoind`). It also loads the Bitcoin wallet by calling the `load_bitcoin_wallet.sh` script.

### `e2e/load_bitcoin_wallet.sh`

Loads the `nativewallet` into the running `bitcoind` container.

### `e2e/init_local_net.sh`

Performs Sui-specific initialization after the `sui-node` is running.  This includes getting the active Sui address and requesting test SUI tokens from the faucet.

### `e2e/deploy_light_client.sh`

Downloads, extracts, deploys, and initializes Bitcoin SPV light client package to the Sui network.  It extracts the Package ID and Light Client ID and updates the `e2e-bitcoin.yml` configuration file.

### `e2e/run_e2e_tests.sh`

Executes the actual end-to-end tests.  This script is a *placeholder*.  You should replace the placeholder `echo` command with the command(s) needed to run E2E tests. All E2E tests should be invoked from this script. The relayer should use the `e2e-bitcoin.yml` file. This file contains the updated package id and object id of the deployed smart contract.

## Workflow

The GitHub Actions workflow (`.github/workflows/e2e-tests.yml`) orchestrates these scripts.

## Important Notes

* The `e2e-bitcoin.yml` should be placed at the root of the project and used for all e2e tests.
