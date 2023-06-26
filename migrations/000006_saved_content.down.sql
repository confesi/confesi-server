BEGIN;

ALTER TABLE saved_posts
    DROP COLUMN updated_at,
    DROP CONSTRAINT unique_saved_post;

ALTER TABLE saved_comments
    DROP COLUMN updated_at,
    DROP CONSTRAINT unique_saved_comment;

END;