-- add auto timestamps to schools table
BEGIN;

ALTER TABLE schools
   DROP COLUMN timezone;

END;