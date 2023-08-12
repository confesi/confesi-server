-- add auto timestamps to schools table
BEGIN;

ALTER TABLE schools
   ADD COLUMN timezone TEXT NOT NULL DEFAULT 'UTC';

END;
