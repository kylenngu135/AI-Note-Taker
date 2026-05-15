CREATE TABLE jobs (
    id          TEXT PRIMARY KEY,
    status      TEXT        NOT NULL DEFAULT 'pending',
    file_key    TEXT        NOT NULL,
    result      TEXT,
    error       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
