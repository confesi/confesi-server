BEGIN;

ALTER TABLE users
    ADD COLUMN room_requests BOOLEAN DEFAULT TRUE NOT NULL;

END;