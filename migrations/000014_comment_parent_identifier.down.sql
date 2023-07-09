BEGIN;

ALTER TABLE comment_identifiers
    DROP COLUMN parent_identifier;

END;