BEGIN;

ALTER TABLE daily_hottest_cron_jobs
    DROP CONSTRAINT daily_hottest_cron_jobs_successfully_ran_unique,
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL,
    ADD COLUMN type VARCHAR(255) NOT NULL;

ALTER TABLE daily_hottest_cron_jobs
    RENAME TO cron_jobs;

ALTER TABLE cron_jobs
    RENAME COLUMN successfully_ran TO ran;

ALTER TABLE users
    DROP COLUMN notifications_enabled;

END;