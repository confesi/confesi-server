BEGIN;

----- comment_identifiers

CREATE TABLE comment_identifiers (
    id SERIAL PRIMARY KEY UNIQUE NOT NULL,
    user_id VARCHAR(255) NOT NULL REFERENCES users (id),
    post_id INTEGER NOT NULL REFERENCES posts (id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    is_op BOOLEAN NOT NULL,
    identifier INTEGER,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_posts FOREIGN KEY (post_id) REFERENCES posts (id),
    CONSTRAINT uq_comment_identifiers UNIQUE (user_id, post_id, is_op, identifier)
);

ALTER TABLE comment_identifiers
    ADD COLUMN parent_identifier INTEGER;

ALTER TABLE comment_identifiers
    DROP CONSTRAINT uq_comment_identifiers;

----- comments

ALTER TABLE comments
    DROP CONSTRAINT fk_comments_root_comment,
    DROP COLUMN root_comment,
    ADD COLUMN ancestors INTEGER[];

ALTER TABLE comments
    ADD COLUMN identifier_id INTEGER,
    ADD CONSTRAINT fk_comment_identifiers FOREIGN KEY (identifier_id) REFERENCES comment_identifiers (id);

ALTER TABLE comments
    DROP COLUMN numerical_user,
    DROP COLUMN numerical_replying_user,
    DROP COLUMN numerical_user_is_op,
    DROP COLUMN numerical_replying_user_is_op;

END;