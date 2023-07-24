BEGIN;

ALTER TABLE posts
    DROP COLUMN year_of_study;

ALTER TABLE posts
    ADD COLUMN year_of_study_id INTEGER REFERENCES year_of_study(id);

END;