BEGIN;

ALTER TABLE votes
    DROP CONSTRAINT IF EXISTS vot_fk_pos,
    DROP CONSTRAINT IF EXISTS vot_fk_com;

ALTER TABLE votes
    ADD FOREIGN KEY (post_id) REFERENCES posts (id),
    ADD FOREIGN KEY (comment_id) REFERENCES comments (id);
    
ALTER TABLE votes
    ADD CONSTRAINT either_post_or_comment CHECK (num_nonnulls(post_id, comment_id) = 1);

END;