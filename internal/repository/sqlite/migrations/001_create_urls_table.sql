CREATE TABLE IF NOT EXISTS urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    short_code TEXT UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    last_used_at DATETIME,
    usage_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);