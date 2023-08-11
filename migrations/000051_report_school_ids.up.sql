BEGIN;

ALTER TABLE reports ADD COLUMN school_id INTEGER; 
ALTER TABLE reports ADD CONSTRAINT sf_fk_sch FOREIGN KEY (school_id) REFERENCES schools (id);

ALTER TABLE comments ADD COLUMN school_id INTEGER; 
ALTER TABLE comments ADD CONSTRAINT sf_fk_sch FOREIGN KEY (school_id) REFERENCES schools (id);

END;