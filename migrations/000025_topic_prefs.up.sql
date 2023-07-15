BEGIN;

ALTER TABLE fcm_topic_prefs
    DROP COLUMN trending_all,
    DROP COLUMN trending_home,
    DROP COLUMN trending_watched,
    ADD COLUMN trending BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN replies_to_your_comments BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN comments_on_your_posts BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN votes_on_your_comments BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN votes_on_your_posts BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN quotes_of_your_posts BOOLEAN NOT NULL DEFAULT TRUE;

END;