BEGIN;

-- change column type back from TEXT to VARCHAR(255) for `title`
ALTER TABLE posts ALTER COLUMN title TYPE VARCHAR(255);

END;
