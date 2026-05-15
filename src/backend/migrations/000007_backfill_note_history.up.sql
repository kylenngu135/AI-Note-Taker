-- Backfill note_history for existing uploads that don't have history entries
-- This creates the initial conversation: user upload (transcription) -> AI response (study sheet)

-- First, ensure all existing history entries have a role (default to 'user' for safety)
UPDATE note_history SET role = 'user' WHERE role IS NULL OR role = '';

-- Insert history entries for uploads that don't have any history yet
INSERT INTO note_history (id, note_id, upload_id, role, prompt, content, storage_key, created_at)
SELECT
    'hist-init-' || u.id,
    n.id,
    u.id,
    'user',
    u.storage_key,  -- We can't recover the original text, so we'll leave this as a placeholder
    '',
    u.storage_key,
    u.created_at
FROM uploads u
JOIN notes n ON n.upload_id = u.id
LEFT JOIN note_history nh ON nh.upload_id = u.id
WHERE nh.id IS NULL;

-- Insert the AI response (study sheet) for these uploads
INSERT INTO note_history (id, note_id, upload_id, role, prompt, content, storage_key, created_at)
SELECT
    'hist-ai-' || u.id,
    n.id,
    u.id,
    'assistant',
    '',
    n.content,
    n.storage_key,
    n.created_at
FROM uploads u
JOIN notes n ON n.upload_id = u.id
LEFT JOIN note_history nh ON nh.upload_id = u.id
WHERE nh.id IS NULL;
