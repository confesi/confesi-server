BEGIN;

ALTER TABLE feedbacks
    DROP CONSTRAINT fk_feedbacks_type;

ALTER TABLE feedbacks
    DROP COLUMN type_id;

DROP TABLE feedback_types;

END;
