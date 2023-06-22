BEGIN;

ALTER TABLE comments
    DROP COLUMN ancestors,
    ADD COLUMN comment_id INT;

END;
