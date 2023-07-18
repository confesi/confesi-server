BEGIN;

ALTER TABLE posts
    DROP COLUMN edited;

ALTER TABLE comments 
    DROP COLUMN edited;

END;