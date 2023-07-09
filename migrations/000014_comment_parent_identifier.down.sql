BEGIN;

ALTER TABLE comment_identifiers
    DROP COLUMN parent_identifier,
    DROP CONSTRAINT u_p_i_i_comment_identifiers;

END;