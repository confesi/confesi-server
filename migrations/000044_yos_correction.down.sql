BEGIN;

ALTER TABLE users
    DROP CONSTRAINT users_yos_fkey;

ALTER TABLE users
    ADD CONSTRAINT users_yos_fkey FOREIGN KEY (year_of_study_id) REFERENCES schools (id);

END;