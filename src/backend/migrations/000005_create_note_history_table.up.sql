CREATE TABLE note_history (
    id              TEXT PRIMARY KEY,
    note_id         TEXT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    upload_id        TEXT NOT NULL REFERENCES uploads(id) ON DELETE CASCADE,
    prompt          TEXT,
    content          TEXT NOT NULL,
    storage_key     TEXT NOT NULL,
    created_at      TIMESTAMP DEFAULT NOW()
);
