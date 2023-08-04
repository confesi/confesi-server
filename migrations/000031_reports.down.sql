BEGIN;

ALTER TABLE reports
    RENAME COLUMN has_been_removed TO handled;

END;