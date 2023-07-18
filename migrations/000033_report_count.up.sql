BEGIN;

ALTER TABLE posts
    ADD COLUMN report_count integer NOT NULL DEFAULT 0,
    ADD COLUMN reviewed_by_mod boolean NOT NULL DEFAULT false;

ALTER TABLE comments
    ADD COLUMN report_count integer NOT NULL DEFAULT 0,
    ADD COLUMN reviewed_by_mod boolean NOT NULL DEFAULT false;

END;