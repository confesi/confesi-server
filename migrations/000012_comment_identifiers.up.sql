BEGIN;

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

ALTER TABLE comments
    ADD COLUMN identifier_id INTEGER REFERENCES comment_identifiers (id);

END;
