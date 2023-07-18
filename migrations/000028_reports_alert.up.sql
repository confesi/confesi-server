BEGIN;

-- First, rename column user_alerted back to handled and change its datatype to bool with default false
ALTER TABLE reports
    RENAME COLUMN user_alerted TO handled;

-- Then, modify the result TEXT column to be nullable
ALTER TABLE reports
    ALTER COLUMN result DROP NOT NULL;

END;
