BEGIN;

ALTER TABLE users
    ADD CONSTRAINT users_yos_fkey FOREIGN KEY (year_of_study_id) REFERENCES schools (id);

END;