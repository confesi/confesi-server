BEGIN;

ALTER TABLE school_follows
    DROP COLUMN created_at,
    DROP COLUMN updated_at;

END;