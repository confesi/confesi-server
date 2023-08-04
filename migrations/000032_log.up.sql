BEGIN;

ALTER TABLE hide_log
    RENAME COLUMN hidden TO removed;

END;