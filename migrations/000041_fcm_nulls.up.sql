BEGIN;

ALTER TABLE fcm_tokens
    ALTER COLUMN user_id DROP NOT NULL;

END;