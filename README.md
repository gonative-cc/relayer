<!-- markdownlint-disable MD041 -->
<!-- markdownlint-disable MD013 -->

<!-- ![Logo!](assets/logo.png) -->

# NATIVE Relayer

[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#wip)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://godoc.org/github.com/gonative-cc/relayer)
[![Go Report Card](https://goreportcard.com/badge/github.com/gonative-cc/relayer?style=flat-square)](https://goreportcard.com/report/github.com/gonative-cc/relayer)
[![Version](https://img.shields.io/github/tag/gonative-cc/relayer.svg?style=flat-square)](https://github.com/gonative-cc/relayer)
[![License: MPL-2.0](https://img.shields.io/github/license/gonative-cc/relayer.svg?style=flat-square)](https://github.com/gonative-cc/relayer/blob/main/LICENSE)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](https://github.com/gonative-cc/contributig/blob/master/CODE_OF_CONDUCT.md)

## Summary

A software that monitors and relayers:

- CometBFT blocks to update Native -> Ika light client
- Bitcoin blocks to update Bitcoin -> Native light client
- Bitcoin SPV proofs to verify dWallet holdings
- Native -> Bitcoin transaction relayer

### Status

Status scheme:

```text
Mock -> WIP -> alpha -> beta -> production
```

| Service          | status |
| :--------------- | :----- |
| Native-\>Ika     | WIP    |
| Bitcoin SPV      |        |
| Native-\>Bitcoin | WIP    |

## Contributing

Participating in open source is often a highly collaborative experience. We’re encouraged to create in public view, and we’re incentivized to welcome contributions of all kinds from people around the world.

Check out [contributing repo](https://github.com/gonative-cc/contributig) for our guidelines & policies for how to contribute. Note: we require DCO! Thank you to all those who have contributed!

### Security

Check out [SECURITY.md](./SECURITY.md) for security concerns.

## Setup

1. Make sure you have `go`, `make` installed
2. Copy and update your env file: `cp .env.example .env`
3. Build the project: `make build`

To build and start you can run: `make build start`

In order to run Native -> Bitcoin relayer PoC:

1. Copy and update .env.example file to .env `cp env.example .env`
2. Run script with path of transaction file as command line argument: `go run main.go transaction.txt`

### Development

1. Run `make setup` (will setup git hooks)
2. Install and make sure it is in your PATH:

   - [markdownlint-cli2](https://github.com/DavidAnson/markdownlint-cli2)
   - [revive](https://github.com/mgechev/revive)

### Coding notes

1. Use `env.Init()` to setup logger and load ENV variables.
1. Use `zerolog.log` logger, eg:

   ```go
   import "github.com/rs/zerolog/log"
   //...
   log.Info().Int("block", minimumBlockHeight).Msg("Start relaying msgs")
   ```

## Talk to us

- Follow the Native team's activities on the [Native X/Twitter account](https://x.com/NativeNetwork).
- Join the conversation on [Native Discord](https://discord.gg/gonative).
