BEGIN;

ALTER TABLE cron_jobs
    RENAME COLUMN ran TO successfully_ran;

ALTER TABLE cron_jobs
    RENAME TO daily_hottest_cron_jobs;

ALTER TABLE daily_hottest_cron_jobs
    DROP COLUMN created_at,
    DROP COLUMN type;

COMMIT;
