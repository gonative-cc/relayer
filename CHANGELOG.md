<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->

<!--
Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
The latest version comes first.
The release date of each version is displayed.

Usage:

Change log entries are to be added to the Unreleased section and in one of the following subsections: Features, Breaking Changes, Bug Fixes. Example entry:

- [#<PR-number>](https://github.com/gonative-cc/relayer/pull/<PR-number>) <description>
-->

# CHANGELOG

## Unreleased

### Features

- Added new functions in the `dal` package:
  - `InsertIkaSignRequest(request IkaSignRequest)`: Inserts a new IKA sign request.
  - `GetIkaSignRequestByID(id uint64)`: Retrieves an IKA sign request by ID.
  - `GetPendingIkaSignRequests()`: Retrieves pending IKA sign requests (no final signature).
  - `UpdateIkaSignRequestFinalSig(id uint64, finalSig Signature)`: Updates the final signature.
  - `GetSignedIkaSignRequests()`: Retrieves signed IKA sign requests not yet broadcasted to Bitcoin.
  - `InsertIkaTx(tx IkaTx)`: Inserts a new IKA transaction.
  - `GetIkaTx(signRequestID uint64, ikaTxID string)`: Retrieves an IKA transaction by primary key.
  - `InsertBitcoinTx(tx BitcoinTx)`: Inserts a new Bitcoin transaction.
  - `GetBitcoinTx(signRequestID uint64, btcTxID []byte)`: Retrieves a Bitcoin transaction by primary key.
  - `GetPendingBitcoinTxs()`: Retrieves Bitcoin transactions with "pending" status.
  - `GetBroadcastedBitcoinTxs()`: Retrieves Bitcoin transactions with "broadcasted" status.
  - `GetBroadcastedBitcoinTxsInfo()`: Retrieves info for "broadcasted" transactions without "confirmed" status.
  - `UpdateBitcoinTxToConfirmed(id uint64, txID []byte)`: Updates Bitcoin transaction to "confirmed" and updates timestamp.

### Breaking Changes

### Bug Fixes

## v0.0.1 (YYYY-MM-DD)
