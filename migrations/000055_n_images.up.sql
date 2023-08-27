BEGIN;

ALTER TABLE posts RENAME COLUMN img_url TO img_urls;
ALTER TABLE posts ALTER COLUMN img_urls TYPE text[] USING CASE WHEN img_urls IS NOT NULL THEN ARRAY[img_urls] ELSE ARRAY[]::text[] END;

END;
