BEGIN;

ALTER TABLE fcm_privs
    DROP COLUMN name,
    ADD COLUMN comment_id INTEGER REFERENCES comments(id),
    ADD COLUMN post_id INTEGER REFERENCES posts(id),    
    ADD CONSTRAINT either_post_or_comment CHECK (num_nonnulls(post_id, comment_id) = 1);

END; 