# Flow

The following diagram illustrates the flow of a transaction through the relayer and the database. This is one `Relayer`, but it interacts with three different networks (`Native`, `Ika` and `Bitcoin`)

```mermaid
flowchart TB
 subgraph subGraph0["Relayer (Native network)"]
    direction TB
        B["Indexes transaction"]
        A["User submits transaction"]
  end
 subgraph subGraph1["Relayer (Ika network)"]
        E["Ika signs transaction"]
        D["Sends to Ika for signing"]
  end
 subgraph subGraph2["Relayer (Bitcoin network)"]
        H["Broadcasts Bitcoin transaction"]
        G["Constructs Bitcoin transaction"]
        I["Checks for confirmations"]
  end
    A --> B
    B -- Stores transaction (Pending) --> Database["Database"]
    Database -- Reads pending transaction --> D
    D --> E
    E -- Stores signed transaction --> Database
    Database -- Reads signed transaction --> G
    G --> H
    H --> I
    I -- Updates status to Confirmed --> Database

    Database@{ shape: db}
```
