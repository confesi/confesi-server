BEGIN;

CREATE TABLE report_types (
    id SERIAL PRIMARY KEY,
    type VARCHAR(255) NOT NULL
);

ALTER TABLE reports
    DROP COLUMN user_id,
    ADD COLUMN comment_id INTEGER REFERENCES comments(id),
    ADD COLUMN post_id INTEGER REFERENCES posts(id),
    ADD CONSTRAINT report_comment_or_post CHECK (
        (comment_id IS NOT NULL AND post_id IS NULL) OR
        (comment_id IS NULL AND post_id IS NOT NULL)
    ),
    DROP COLUMN type,
    ADD COLUMN type_id INTEGER REFERENCES report_types(id),
    ADD COLUMN updated_at TIMESTAMPTZ;

END;