BEGIN;

CREATE TABLE year_of_study (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);


ALTER TABLE users
    RENAME COLUMN year_of_study TO year_of_study_id;

ALTER TABLE users
    DROP COLUMN mod_id,
    ADD COLUMN is_limited BOOLEAN NOT NULL DEFAULT FALSE;


DROP TABLE IF EXISTS mod_levels;

END;