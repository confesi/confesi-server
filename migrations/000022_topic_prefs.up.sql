BEGIN;

CREATE TABLE fcm_topic_prefs (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL
        REFERENCES users(id),
    daily_hottest  BOOLEAN NOT NULL DEFAULT TRUE,
    trending_all BOOLEAN NOT NULL DEFAULT TRUE,
    trending_home  BOOLEAN NOT NULL DEFAULT TRUE,
    trending_watched  BOOLEAN NOT NULL DEFAULT TRUE,
    new_features BOOLEAN NOT NULL DEFAULT TRUE
);

END;