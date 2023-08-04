BEGIN;

ALTER TABLE fcm_tokens
    DROP CONSTRAINT fcm_tokens_user_id_fkey;

ALTER TABLE fcm_tokens
    ADD CONSTRAINT fcm_tokens_user_id_token_unique UNIQUE (user_id, token);

END;
