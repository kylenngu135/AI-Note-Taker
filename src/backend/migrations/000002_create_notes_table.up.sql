CREATE TABLE notes (
    id              TEXT PRIMARY KEY,
    upload_id       TEXT NOT NULL REFERENCES uploads(id) ON DELETE CASCADE,
    content         TEXT NOT NULL,
    storage_key     TEXT NOT NULL,
    created_at      TIMESTAMP DEFAULT NOW(),
    last_updated_at TIMESTAMP DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION update_notes_last_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notes_last_updated_at
    BEFORE UPDATE ON notes
    FOR EACH ROW
    EXECUTE FUNCTION update_notes_last_updated_at();
