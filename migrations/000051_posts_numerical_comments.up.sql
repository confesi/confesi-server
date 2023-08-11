BEGIN;

ALTER TABLE posts ADD COLUMN comment_numerics INTEGER NOT NULL DEFAULT 0; 
-- Make down_vote and up_vote contrainted to be positive on posts table
ALTER TABLE posts ADD CONSTRAINT upvotes_positive CHECK (upvote >= 0);
ALTER TABLE posts ADD CONSTRAINT downvotes_positive CHECK (downvote >= 0);

END;

