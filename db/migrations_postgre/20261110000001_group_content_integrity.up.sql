BEGIN;

-- Block writes while installing the guards and repairing historical rows.
LOCK TABLE kuis, cerita_interaktif, puzzles, group_contents IN SHARE ROW EXCLUSIVE MODE;

CREATE OR REPLACE FUNCTION validate_group_content_target() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    -- The row lock prevents a concurrent DELETE from passing this INSERT.
    CASE NEW.content_type::text
        WHEN 'kuis' THEN
            PERFORM 1 FROM kuis WHERE kuis_id = NEW.content_id FOR KEY SHARE;
        WHEN 'cerita' THEN
            PERFORM 1 FROM cerita_interaktif WHERE cerita_id = NEW.content_id FOR KEY SHARE;
        WHEN 'puzzle' THEN
            PERFORM 1 FROM puzzles WHERE puzzle_id = NEW.content_id FOR KEY SHARE;
        ELSE
            RAISE EXCEPTION 'Unsupported group content type: %', NEW.content_type USING ERRCODE = '23503';
    END CASE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Group content target does not exist: %/%', NEW.content_type, NEW.content_id
            USING ERRCODE = '23503';
    END IF;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION delete_group_content_relations() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM group_contents
    WHERE content_type = TG_ARGV[0]::content_type_enum
      AND content_id = (to_jsonb(OLD)->>TG_ARGV[1])::integer;
    RETURN OLD;
END;
$$;

CREATE TRIGGER group_content_target_guard
BEFORE INSERT OR UPDATE OF content_type, content_id ON group_contents
FOR EACH ROW EXECUTE FUNCTION validate_group_content_target();

CREATE TRIGGER kuis_delete_group_contents AFTER DELETE ON kuis
FOR EACH ROW EXECUTE FUNCTION delete_group_content_relations('kuis', 'kuis_id');
CREATE TRIGGER cerita_delete_group_contents AFTER DELETE ON cerita_interaktif
FOR EACH ROW EXECUTE FUNCTION delete_group_content_relations('cerita', 'cerita_id');
CREATE TRIGGER puzzle_delete_group_contents AFTER DELETE ON puzzles
FOR EACH ROW EXECUTE FUNCTION delete_group_content_relations('puzzle', 'puzzle_id');

CREATE INDEX IF NOT EXISTS group_contents_target_idx ON group_contents (content_type, content_id);

DELETE FROM group_contents gc
WHERE (content_type = 'kuis' AND NOT EXISTS (SELECT 1 FROM kuis WHERE kuis_id = gc.content_id))
   OR (content_type = 'cerita' AND NOT EXISTS (SELECT 1 FROM cerita_interaktif WHERE cerita_id = gc.content_id))
   OR (content_type = 'puzzle' AND NOT EXISTS (SELECT 1 FROM puzzles WHERE puzzle_id = gc.content_id));

COMMIT;
