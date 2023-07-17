BEGIN;

ALTER TABLE reports
    RENAME COLUMN handled TO has_been_removed;

END;