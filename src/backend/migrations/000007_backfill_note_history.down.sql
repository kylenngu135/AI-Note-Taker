-- Remove backfilled history entries
DELETE FROM note_history WHERE id LIKE 'hist-init-%' OR id LIKE 'hist-ai-%';
