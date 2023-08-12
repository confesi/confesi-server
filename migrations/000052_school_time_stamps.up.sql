-- add auto timestamps to schools table
BEGIN;

ALTER TABLE schools
   ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('UTC'::text, now()),
   ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT timezone('UTC'::text, now());

END;
