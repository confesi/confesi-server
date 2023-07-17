BEGIN;

ALTER TABLE reports
    DROP CONSTRAINT reports_user_comment_id_post_id_unique;

END;