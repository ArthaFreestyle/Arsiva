ALTER TABLE cerita_interaktif ALTER COLUMN created_at DROP NOT NULL;
ALTER TABLE cerita_interaktif ALTER COLUMN created_at DROP DEFAULT;
-- Backfilled values are intentionally kept: the original NULLs carry no
-- information and cannot be recovered.
