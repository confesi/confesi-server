BEGIN;

ALTER TABLE posts DROP COLUMN IF EXISTS comment_numerics;
-- Remove Positive constraint on up_vote and down_vote
-- Drop the constraint using its name
ALTER TABLE posts DROP CONSTRAINT IF EXISTS upvotes_positive;
ALTER TABLE posts DROP CONSTRAINT IF EXISTS downvotes_positive;

END;
