BEGIN;

ALTER TABLE schools
    DROP CONSTRAINT unique_school_names;

ALTER TABLE feedback_types
    DROP CONSTRAINT unique_feedback_types;

ALTER TABLE report_types
    DROP CONSTRAINT unique_report_types;

ALTER TABLE post_categories
    DROP CONSTRAINT unique_post_categories;


END;
