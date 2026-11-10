BEGIN;
DROP TRIGGER IF EXISTS group_content_target_guard ON group_contents;
DROP TRIGGER IF EXISTS kuis_delete_group_contents ON kuis;
DROP TRIGGER IF EXISTS cerita_delete_group_contents ON cerita_interaktif;
DROP TRIGGER IF EXISTS puzzle_delete_group_contents ON puzzles;
DROP FUNCTION IF EXISTS validate_group_content_target();
DROP FUNCTION IF EXISTS delete_group_content_relations();
DROP INDEX IF EXISTS group_contents_target_idx;
-- Deleted historical orphan relations cannot be restored by a rollback.
COMMIT;
