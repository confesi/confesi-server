BEGIN;

ALTER TABLE fcm_tokens
    ALTER COLUMN user_id SET NOT NULL;

END;