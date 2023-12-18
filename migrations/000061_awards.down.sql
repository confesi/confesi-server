BEGIN;

-- Drop the awards table first because it has foreign key constraints
-- that reference the award_types table.
DROP TABLE IF EXISTS awards;

-- Drop the award_types table.
DROP TABLE IF EXISTS award_types;

END;
