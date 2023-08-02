BEGIN;

CREATE TABLE feedback_types (
    id SERIAL PRIMARY KEY,
    type TEXT
);

ALTER TABLE feedbacks
    ADD COLUMN type_id INT,
    ADD CONSTRAINT fk_feedbacks_type
    FOREIGN KEY (type_id)
    REFERENCES feedback_types (id);

END;
