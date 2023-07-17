BEGIN;

CREATE TABLE hide_log (
    id SERIAL PRIMARY KEY UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    comment_id INTEGER REFERENCES comments(id),
    post_id INTEGER REFERENCES posts(id),
    reason TEXT DEFAULT NULL,
    hidden BOOLEAN NOT NULL,
    user_id VARCHAR(255) REFERENCES users(id),
    CONSTRAINT report_comment_or_post CHECK (
        (comment_id IS NOT NULL AND post_id IS NULL) OR
        (comment_id IS NULL AND post_id IS NOT NULL)
    )
);

END;
