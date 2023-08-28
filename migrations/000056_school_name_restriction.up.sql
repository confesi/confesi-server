BEGIN;

ALTER TABLE schools
    ADD CONSTRAINT unique_school_names UNIQUE (name);
ALTER TABLE feedback_types
    ADD CONSTRAINT unique_feedback_types UNIQUE (type);

ALTER TABLE report_types
    ADD CONSTRAINT unique_report_types UNIQUE (type);

ALTER TABLE post_categories
    ADD CONSTRAINT unique_post_categories UNIQUE (name);
END;
