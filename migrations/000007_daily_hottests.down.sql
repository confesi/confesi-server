BEGIN;

    DROP INDEX IF EXISTS idx_posts_hottest_on;
    DROP INDEX IF EXISTS idx_posts_trending_score;

    ALTER TABLE schools DROP COLUMN daily_hottests;
    DROP TABLE IF EXISTS daily_hottest_cron_jobs;

END;