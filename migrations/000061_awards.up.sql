BEGIN;

CREATE TABLE award_types (
    id SERIAL PRIMARY KEY UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    icon VARCHAR(255) NOT NULL
);

CREATE TABLE awards_total (
    id SERIAL PRIMARY KEY UNIQUE NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    award_type_id INTEGER NOT NULL,
    total INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (award_type_id) REFERENCES award_types(id)
);

CREATE TABLE awards_general (
    id SERIAL PRIMARY KEY UNIQUE NOT NULL,
    post_id INTEGER,
    comment_id INTEGER,
    user_id VARCHAR(255) NOT NULL,
    award_type_id INTEGER NOT NULL,
    FOREIGN KEY (post_id) REFERENCES posts(id),
    FOREIGN KEY (comment_id) REFERENCES comments(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (award_type_id) REFERENCES award_types(id),
    CHECK ((post_id IS NOT NULL AND comment_id IS NULL) OR (post_id IS NULL AND comment_id IS NOT NULL))
);

CREATE INDEX idx_awards_total_user_id ON awards_total(user_id);
CREATE INDEX idx_awards_general_user_id ON awards_general(user_id);

END;
