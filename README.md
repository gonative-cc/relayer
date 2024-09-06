<!-- markdownlint-disable MD041 -->
<!-- markdownlint-disable MD013 -->

<!-- ![Logo!](assets/logo.png) -->

# NATIVE Relayer

[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#wip)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://godoc.org/github.com/gonative-cc/relayer)
[![Go Report Card](https://goreportcard.com/badge/github.com/gonative-cc/relayer?style=flat-square)](https://goreportcard.com/report/github.com/gonative-cc/relayer)
[![Version](https://img.shields.io/github/tag/gonative-cc/relayer.svg?style=flat-square)](https://github.com/github.com/gonative-cc/relayer)
[![License: MPL-2.0](https://img.shields.io/github/license/gonative-cc/relayer.svg?style=flat-square)](https://github.com/gonative-cc/relayer/blob/main/LICENSE)

## Summary

A software that monitors and relayers:

- CometBFT blocks to update Native -> Pera light client
- Bitcoin blocks to update Bitcoin -> Native light client
- Bitcoin SPV proofs to verify dWallet holdings

### Status

Status scheme:

```text
Mock -> WIP -> alpha -> beta -> production
```

| Service      | status |
| :----------- | :----- |
| Native->Pera | WIP    |
| Bitcoin SPV  |        |

## Contributing

Check out [CONTRIBUTING.md](./CONTRIBUTING.md) for our guidelines & policies for how we develop the Cosmos Hub. Thank you to all those who have contributed!
