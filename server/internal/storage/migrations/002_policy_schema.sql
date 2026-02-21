CREATE TABLE IF NOT EXISTS policies (
    wallet_id TEXT PRIMARY KEY,
    single_tx_limit TEXT,
    daily_limit TEXT,
    whitelist TEXT,
    daily_tx_limit INTEGER NOT NULL DEFAULT 0,
    start_time INTEGER,
    end_time INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS daily_usage (
    wallet_id TEXT NOT NULL,
    date TEXT NOT NULL,
    total_amount TEXT NOT NULL,
    tx_count INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (wallet_id, date)
);

CREATE INDEX IF NOT EXISTS idx_daily_usage_wallet_date ON daily_usage(wallet_id, date);
