BEGIN;

ALTER TABLE posts DROP COLUMN comment_numerics;
-- Remove Postive constraint on up_vote and down_vote
ALTER TABLE posts DROP CONSTRAINT upvotes_positive;
ALTER TABLE posts DROP CONSTRAINT downvotes_positive;

END;