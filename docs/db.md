# Database Interactions

This part describes the database schema and interactions used in the relayer.

## Schema

The relayer uses a SQLite database to store and manage transaction data. The database has a single table named `transactions` with the following schema:

| Column | Data Type | Description |
|---|---|---|
| `btc_tx_id` | INTEGER | The ID of the Bitcoin transaction (primary key) |
| `raw_tx` | BLOB | The raw transaction data as a byte array |
| `tx_hash` | BLOB | The transaction hash as a byte array |
| `status` | INTEGER | The status of the transaction (0: Pending, 1: Signed, 2: Broadcasted, 3: Confirmed) |

## Functions

The relayer interacts with the database using the following functions in the `dal` package:

1. `InsertTx()`: Inserts a new transaction into the database.
2. `GetTx()`: Retrieves a transaction by its ID.
3. `GetSignedTxs()`: Retrieves all transactions with a `signed` status.
4. `GetBroadcastedTxs()`: Retrives all transactions with a `broadcasted` status.
5. `UpdateTxStatus()`: Updates the status of a transaction.
