BEGIN;

ALTER TABLE comments
    ADD COLUMN numerical_user INTEGER,
    ADD COLUMN numerical_replying_user INTEGER,
    ADD COLUMN numerical_user_is_op BOOLEAN,
    ADD COLUMN numerical_replying_user_is_op BOOLEAN,
    DROP COLUMN ancestors,
    ADD COLUMN parent_root INTEGER,
    ADD CONSTRAINT fk_comments_root_comment FOREIGN KEY (parent_root) REFERENCES comments (id),
    DROP COLUMN identifier_id;

DROP TABLE comment_identifiers;

END;