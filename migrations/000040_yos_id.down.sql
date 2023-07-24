BEGIN;

ALTER TABLE posts
    DROP COLUMN year_of_study_id;

ALTER TABLE posts
    ADD COLUMN year_of_study INTEGER;

END;
