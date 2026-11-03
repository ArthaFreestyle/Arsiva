-- cerita_interaktif.created_at was declared without a DEFAULT and the API insert
-- omits it, so every story created through POST /v1/stories has created_at NULL
-- (only seeded rows have a value). Backfill, then make the column behave like
-- created_at on kuis / puzzles / artikel.

-- Best available proxy for when a story was created: its thumbnail upload time.
UPDATE cerita_interaktif c
SET created_at = a.created_at
FROM assets a
WHERE c.created_at IS NULL AND a.asset_id = c.thumbnail_asset_id;

UPDATE cerita_interaktif SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;

ALTER TABLE cerita_interaktif ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE cerita_interaktif ALTER COLUMN created_at SET NOT NULL;
