BEGIN;

ALTER TABLE cron_jobs
    RENAME COLUMN ran TO successfully_ran;

ALTER TABLE cron_jobs
    RENAME TO daily_hottest_cron_jobs;

ALTER TABLE daily_hottest_cron_jobs
    ADD CONSTRAINT daily_hottest_cron_jobs_successfully_ran_unique UNIQUE (successfully_ran),
    DROP COLUMN created_at,
    DROP COLUMN type;

ALTER TABLE users
    ADD COLUMN notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE;

END;
