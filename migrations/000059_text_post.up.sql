BEGIN;

ALTER TABLE posts
    ADD COLUMN chat_post BOOLEAN NOT NULL DEFAULT FALSE;

END;