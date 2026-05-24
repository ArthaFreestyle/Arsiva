-- Backfill level for all existing members using the same formula as LevelForXP:
-- level = floor(sqrt(total_xp / 100.0))
UPDATE members SET level = FLOOR(SQRT(total_xp / 100.0));
