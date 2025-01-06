# Bitcoin SPV relayer

The code is adapted from [https://github.com/babylonchain/vigilante/tree/dev/reporter](https://github.com/babylonchain/vigilante/tree/dev/reporter).

This relayer is responsible for:

- syncing the latest BTC blocks with a BTC node
- detecting and reporting inconsistency between BTC blockchain and Babylon BTCLightclient header chain