BEGIN;

ALTER TABLE schools ADD COLUMN website VARCHAR(255) NOT NULL DEFAULT '';

END;