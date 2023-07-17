BEGIN;

-- Add a unique constraint for (reported_by, comment_id)
ALTER TABLE reports ADD CONSTRAINT idx_reports_comments UNIQUE (reported_by, comment_id);

-- Add a unique constraint for (reported_by, post_id)
ALTER TABLE reports ADD CONSTRAINT idx_reports_posts UNIQUE (reported_by, post_id);

END;
