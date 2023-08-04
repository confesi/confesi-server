BEGIN;

ALTER TABLE saved_posts
    ADD CONSTRAINT unique_saved_post UNIQUE (user_id, post_id),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL;

ALTER TABLE saved_comments
    ADD CONSTRAINT unique_saved_comment UNIQUE (user_id, comment_id),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL;

END;