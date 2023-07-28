BEGIN;

ALTER TABLE fcm_tokens
DROP CONSTRAINT fcm_tokens_user_id_token_unique;

ALTER TABLE fcm_tokens
ADD CONSTRAINT fcm_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id);

END;
