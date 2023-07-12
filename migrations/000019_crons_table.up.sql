BEGIN;

ALTER TABLE daily_hottest_cron_jobs
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL,
    ADD COLUMN type VARCHAR(255) NOT NULL;

ALTER TABLE daily_hottest_cron_jobs
    RENAME TO cron_jobs;

ALTER TABLE cron_jobs
    RENAME COLUMN successfully_ran TO ran;

END;