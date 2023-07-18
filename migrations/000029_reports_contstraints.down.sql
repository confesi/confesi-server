BEGIN;

ALTER TABLE reports
    DROP CONSTRAINT IF EXISTS idx_reports_comments,
    DROP CONSTRAINT IF EXISTS idx_reports_posts;

END;