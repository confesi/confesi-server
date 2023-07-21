BEGIN;

ALTER TABLE posts
    ADD COLUMN year_of_study INTEGER,
    ADD CONSTRAINT posts_year_of_study_fkey FOREIGN KEY (year_of_study) REFERENCES year_of_study (id);

END;