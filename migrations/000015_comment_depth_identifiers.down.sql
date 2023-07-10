BEGIN;

ALTER TABLE comment_identifiers
    ADD CONSTRAINT uq_comment_identifiers UNIQUE (user_id, post_id, is_op, identifier);

END;