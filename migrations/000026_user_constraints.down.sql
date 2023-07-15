BEGIN;

ALTER TABLE fcm_topic_prefs
    DROP CONSTRAINT fcm_topic_prefs_user_id_unique;

END;
