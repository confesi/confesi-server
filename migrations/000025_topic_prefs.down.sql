BEGIN;

ALTER TABLE fcm_topic_prefs
    ADD COLUMN trending_all BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN trending_home BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN trending_watched BOOLEAN NOT NULL DEFAULT TRUE,
    DROP COLUMN trending,
    DROP COLUMN replies_to_your_comments,
    DROP COLUMN comments_on_your_posts,
    DROP COLUMN votes_on_your_comments,
    DROP COLUMN votes_on_your_posts,
    DROP COLUMN quotes_of_your_posts;

END;