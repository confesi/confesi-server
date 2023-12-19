BEGIN;

-- Drop the indexes
DROP INDEX IF EXISTS idx_awards_total_user_id;
DROP INDEX IF EXISTS idx_awards_general_user_id;

-- Drop the tables
DROP TABLE IF EXISTS awards_general;
DROP TABLE IF EXISTS awards_total;
DROP TABLE IF EXISTS award_types;

END;
