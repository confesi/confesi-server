BEGIN;

    CREATE TABLE daily_hottest_cron_jobs (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        successfully_ran DATE,
        CONSTRAINT daily_hottest_cron_jobs_successfully_ran_unique UNIQUE (successfully_ran)
    );

    ALTER TABLE schools ADD COLUMN daily_hottests INTEGER DEFAULT 0;

END;
