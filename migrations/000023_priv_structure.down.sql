BEGIN;

ALTER TABLE fcm_privs
    ADD COLUMN name VARCHAR(255) NOT NULL,
    DROP CONSTRAINT IF EXISTS either_post_or_comment,
    DROP COLUMN comment_id,
    DROP COLUMN post_id;


END;