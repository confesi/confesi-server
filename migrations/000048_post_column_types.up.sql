BEGIN;

-- change column type form VARCHAR(255) to TEXT for `title`
ALTER TABLE posts ALTER COLUMN title TYPE TEXT;

END;
