BEGIN;

ALTER TABLE posts
    DROP COLUMN report_count,
    DROP COLUMN reviewed_by_mod;

ALTER TABLE comments
    DROP COLUMN report_count,
    DROP COLUMN reviewed_by_mod;

END;