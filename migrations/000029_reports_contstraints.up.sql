BEGIN;

ALTER TABLE reports
    ADD CONSTRAINT reports_user_comment_id_post_id_unique UNIQUE (reported_by, comment_id, post_id);

END;