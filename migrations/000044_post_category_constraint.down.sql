BEGIN;

ALTER TABLE posts ALTER COLUMN category_id DROP NOT NULL;

END;