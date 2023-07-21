BEGIN;

ALTER TABLE users
    DROP CONSTRAINT users_yos_fkey;

END;