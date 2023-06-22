BEGIN;

CREATE INDEX posts_trending_score_idx ON posts (trending_score);
CREATE INDEX posts_created_at_idx ON posts (created_at);

END;