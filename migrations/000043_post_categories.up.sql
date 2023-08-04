BEGIN;


CREATE TABLE post_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

ALTER TABLE posts ADD COLUMN category_id INTEGER REFERENCES post_categories(id);

END;
