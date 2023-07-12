BEGIN;

ALTER TABLE fcm_tokens RENAME TO notifications;

ALTER TABLE fcm_topics RENAME TO topics;

END;
