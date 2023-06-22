BEGIN;

ALTER TABLE comments
    DROP COLUMN comment_id,
    ADD COLUMN ancestors INT[];

END;