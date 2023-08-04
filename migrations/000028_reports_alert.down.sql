BEGIN;

ALTER TABLE reports
    DROP COLUMN result,
     ADD COLUMN result TEXT;


ALTER TABLE reports
    RENAME COLUMN handled TO user_alerted;

COMMIT;