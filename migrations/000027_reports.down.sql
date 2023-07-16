BEGIN;

ALTER TABLE reports
    ADD COLUMN user_id VARCHAR(255) NOT NULL,
    ADD CONSTRAINT rep_fk_use FOREIGN KEY (user_id) REFERENCES users (id),
    DROP CONSTRAINT report_comment_or_post,
    DROP COLUMN comment_id,
    DROP COLUMN post_id,
    ADD COLUMN type VARCHAR(255) NOT NULL,
    DROP COLUMN updated_at,
    DROP COLUMN type_id;

DROP TABLE report_types;

END;