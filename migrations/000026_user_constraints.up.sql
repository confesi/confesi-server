BEGIN;

ALTER TABLE fcm_topic_prefs
    ADD CONSTRAINT fcm_topic_prefs_user_id_unique UNIQUE (user_id);

END;