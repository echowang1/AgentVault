CREATE TABLE IF NOT EXISTS wallets (
    id TEXT PRIMARY KEY,
    address TEXT UNIQUE NOT NULL,
    public_key TEXT NOT NULL,
    shard2_id TEXT UNIQUE NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_wallets_address ON wallets(address);
CREATE INDEX IF NOT EXISTS idx_wallets_shard2_id ON wallets(shard2_id);

CREATE TABLE IF NOT EXISTS key_shards (
    id TEXT PRIMARY KEY,
    shard2_encrypted BLOB NOT NULL,
    nonce BLOB NOT NULL,
    created_at INTEGER NOT NULL
);
