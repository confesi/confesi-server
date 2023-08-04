BEGIN;

ALTER TABLE notifications RENAME TO fcm_tokens;

ALTER TABLE topics RENAME TO fcm_topics;

END;
