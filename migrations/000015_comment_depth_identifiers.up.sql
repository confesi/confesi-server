BEGIN;

ALTER TABLE comment_identifiers
    DROP CONSTRAINT uq_comment_identifiers;

END;