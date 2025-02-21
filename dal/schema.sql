CREATE TABLE IF NOT EXISTS ika_sign_requests (
    id INTEGER PRIMARY KEY,
    payload BLOB NOT NULL,
    dWallet_id TEXT NOT NULL,
    user_sig TEXT NOT NULL,
    final_sig BLOB,
    timestamp INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS ika_txs (
    sr_id INTEGER NOT NULL,  -- sign request
    status INTEGER NOT NULL,
    ika_tx_id TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    note TEXT,
    PRIMARY KEY (sr_id, ika_tx_id),
    FOREIGN KEY (sr_id) REFERENCES ika_sign_requests (id)
);

CREATE TABLE IF NOT EXISTS bitcoin_txs (
    sr_id INTEGER NOT NULL,
    status INTEGER NOT NULL,
    btc_tx_id BLOB NOT NULL,
    timestamp INTEGER NOT NULL,
    note TEXT,
    PRIMARY KEY (sr_id, btc_tx_id),
    FOREIGN KEY (sr_id) REFERENCES ika_sign_requests (id)
);