BEGIN;

-- Remove the CHECK constraint
ALTER TABLE awards_general DROP CONSTRAINT IF EXISTS awards_general_post_id_comment_id_check;

END;
