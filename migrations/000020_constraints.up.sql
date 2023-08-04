BEGIN;

ALTER TABLE fcm_tokens ADD CONSTRAINT fcm_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id);

ALTER TABLE fcm_topics ADD CONSTRAINT fcm_topics_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id);

END;