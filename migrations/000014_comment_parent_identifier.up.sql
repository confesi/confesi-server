BEGIN;

ALTER TABLE comment_identifiers
    ADD COLUMN parent_identifier INTEGER,
    ADD CONSTRAINT u_p_i_i_comment_identifiers UNIQUE (user_id, post_id, is_op, identifier, parent_identifier);

END;