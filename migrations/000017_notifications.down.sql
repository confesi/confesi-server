BEGIN;

DROP TABLE notifications;

DROP TABLE topics;

ALTER TABLE users
    DROP COLUMN notifications_enabled;

END;