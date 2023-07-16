BEGIN;

CREATE TABLE topics (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT topic_fk_use 
        FOREIGN KEY (user_id)
        REFERENCES users (id)
);

ALTER TABLE topics RENAME TO fcm_topics;

ALTER TABLE fcm_topics ADD CONSTRAINT fcm_topics_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id);

ALTER TABLE fcm_topics RENAME TO fcm_privs;

ALTER TABLE fcm_privs
    DROP COLUMN name,
    ADD COLUMN comment_id INTEGER REFERENCES comments(id),
    ADD COLUMN post_id INTEGER REFERENCES posts(id),    
    ADD CONSTRAINT either_post_or_comment CHECK (num_nonnulls(post_id, comment_id) = 1);

END;