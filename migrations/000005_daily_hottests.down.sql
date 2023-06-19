BEGIN;

    ALTER TABLE schools DROP COLUMN daily_hottests;
    DROP TABLE IF EXISTS daily_hottest_cron_jobs;

END;