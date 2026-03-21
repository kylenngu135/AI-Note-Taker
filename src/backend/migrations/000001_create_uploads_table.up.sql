CREATE TABLE uploads (
    id              TEXT PRIMARY KEY,
    filename        TEXT NOT NULL,
    file_type       TEXT NOT NULL,
    file_size       BIGINT NOT NULL,
    storage_key     TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'complete',
    created_at      TIMESTAMP DEFAULT NOW(),
    last_updated_at TIMESTAMP DEFAULT NOW()
);
