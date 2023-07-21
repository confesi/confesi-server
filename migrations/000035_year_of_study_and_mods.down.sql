BEGIN;


ALTER TABLE users
    RENAME COLUMN year_of_study_id TO year_of_study;

ALTER TABLE users
    ADD COLUMN mod_id INTEGER,
    DROP COLUMN is_limited;

DROP TABLE IF EXISTS year_of_study;

CREATE TABLE mod_levels (
    id INTEGER PRIMARY KEY UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL
);

END;
