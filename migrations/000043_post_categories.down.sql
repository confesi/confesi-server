BEGIN;

ALTER TABLE posts DROP COLUMN category_id;

DROP TABLE post_categories;

END;