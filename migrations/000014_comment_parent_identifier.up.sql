BEGIN;

ALTER TABLE comment_identifiers
    ADD COLUMN parent_identifier INTEGER;

END;