-- add auto timestamps to schools table
BEGIN;

ALTER TABLE schools
   DROP COLUMN created_at,
  DROP COLUMN updated_at;

END;