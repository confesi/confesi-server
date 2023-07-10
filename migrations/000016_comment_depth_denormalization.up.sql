BEGIN;

ALTER TABLE comments
    ADD COLUMN numerical_user INTEGER,
    ADD COLUMN numerical_replying_user INTEGER,
    ADD COLUMN numerical_user_is_op INTEGER,
    ADD COLUMN numerical_replying_user_is_op INTEGER,
    DROP COLUMN ancestors,
    ADD COLUMN root_comment INTEGER,
    ADD CONSTRAINT fk_comments_root_comment FOREIGN KEY (root_comment) REFERENCES comments (id),
    DROP COLUMN identifier_id;

DROP TABLE comment_identifiers;

END;