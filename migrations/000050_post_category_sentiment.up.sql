BEGIN;

ALTER TABLE posts ADD COLUMN sentiment FLOAT NOT NULL DEFAULT 0.0; 

END;
