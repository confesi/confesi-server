BEGIN;

ALTER TABLE comments
    DROP COLUMN ancestors,
    DROP COLUMN children_count,
    ADD COLUMN comment_id INT;

END;
