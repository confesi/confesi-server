BEGIN;

ALTER TABLE posts   
    ADD COLUMN img_url TEXT;

END;