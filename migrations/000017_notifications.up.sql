BEGIN;

CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    token VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    user_id VARCHAR(255) NOT NULL
);

CREATE TABLE topics (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT topic_fk_use 
        FOREIGN KEY (user_id)
        REFERENCES users (id)
);

ALTER TABLE users
    ADD COLUMN notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE;

END;