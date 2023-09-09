-- Create a notification logging table
-- add primary key of user id referencing users table

CREATE TABLE notification_logs (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255),
    body TEXT,
    data TEXT,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    read BOOLEAN NOT NULL DEFAULT FALSE
);