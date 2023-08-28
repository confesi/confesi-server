BEGIN;

ALTER TABLE posts RENAME COLUMN img_urls TO img_url;
ALTER TABLE posts ALTER COLUMN img_url TYPE text USING COALESCE(img_url[1], NULL);

END;
