BEGIN;

ALTER TABLE hide_log
    RENAME COLUMN removed TO hidden;

END;