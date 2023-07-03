BEGIN;

ALTER TABLE comments
    DROP COLUMN comment_id,
    ADD COLUMN children_count INT,
    ADD COLUMN ancestors INT[];

END;