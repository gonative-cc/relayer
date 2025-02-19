# Bitcoin-SPV Flow

Following diagram explains how the bitcoin-SPV relayer interacts with `BitcoinNode` and `LightClient` and how data flows from Bitcoin node to Lightclient through the SPV relayer.

## Connecting

```mermaid
flowchart TD
    A["Bitcoin full node"] <-- 1 Connect --> B("Native SPV relayer")
    B <-- 2 Connect --> D["SPV lightclient"]
    B -- 3 Sync blocks --> D
    B <-. 4 New events listen .-> A
```

## Sending block headers

```mermaid
flowchart TD
    A["Bitcoin full node"] -. 1 New blocks .-> B("Native SPV relayer")
    B -- 2 Send blockheader --> D["SPV lightclient"]
```

## Sending SPV proofs

```mermaid
flowchart TD
    A["Bitcoin full node"] -. 1 New transaction .-> B("Native SPV relayer")
    B -- 2 Send SPV proof --> D["SPV lightclient"]
    D -- 3 Valid/Invalid --> B
```
