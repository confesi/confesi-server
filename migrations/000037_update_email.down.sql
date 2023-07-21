BEGIN;

ALTER TABLE users
    DROP COLUMN email_updated_at;

END;