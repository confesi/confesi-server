BEGIN;

ALTER TABLE users
    DROP COLUMN room_requests;

END;