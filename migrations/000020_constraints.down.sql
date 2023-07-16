BEGIN;

ALTER TABLE fcm_tokens DROP CONSTRAINT fcm_tokens_user_id_fkey;

ALTER TABLE fcm_topics DROP CONSTRAINT fcm_topics_user_id_fkey;

END;