BEGIN;

ALTER TABLE votes
    DROP CONSTRAINT IF EXISTS either_post_or_comment;

ALTER TABLE votes
    ADD CONSTRAINT vot_fk_pos FOREIGN KEY (post_id) REFERENCES posts (id),
    ADD CONSTRAINT vot_fk_com FOREIGN KEY (comment_id) REFERENCES comments (id);

END;
