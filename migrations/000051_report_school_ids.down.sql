BEGIN;

ALTER TABLE reports DROP CONSTRAINT sf_fk_sch;
ALTER TABLE reports DROP COLUMN school_id;

ALTER TABLE comments DROP CONSTRAINT sf_fk_sch;
ALTER TABLE comments DROP COLUMN school_id;

END;