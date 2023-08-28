-- Create Indexes on columns that are used in WHERE clauses
-- TABLE                          COLUMN
-- ---------------------------------------
-- comments                       hidden
-- comments                       user_id
-- comments                       post_id
-- comments                       created_at
-- comments                       trending_score
-- users                          school_id
-- drafts                         user_id
-- fcm_tokens                     user_id
-- fcm_topic_prefs                user_id
-- posts                          school_id
-- posts                          user_id
-- posts                          created_at 
-- posts                          trending_score
-- posts                          sentiment                      
-- reports                        created_at
-- reports                        reported_by
-- saved_comments                 user_id
-- saved_comments                 comment_id
-- saved_posts                    post_id
-- saved_posts                    user_id
-- school_follows                 user_id
-- votes                          post_id             
-- votes                          comment_id
BEGIN;

CREATE INDEX comments_hidden_idx ON comments (hidden);
CREATE INDEX comments_user_id_idx ON comments (user_id);
CREATE INDEX comments_post_id_idx ON comments (post_id);
CREATE INDEX comments_created_at_idx ON comments (created_at);
CREATE INDEX comments_trending_score_idx ON comments (trending_score);
CREATE INDEX users_school_id_idx ON users (school_id);
CREATE INDEX drafts_user_id_idx ON drafts (user_id);
CREATE INDEX fcm_tokens_user_id_idx ON fcm_tokens (user_id);
CREATE INDEX fcm_topic_prefs_user_id_idx ON fcm_topic_prefs (user_id);
CREATE INDEX posts_school_id_idx ON posts (school_id);
CREATE INDEX posts_user_id_idx ON posts (user_id);
CREATE INDEX posts_sentiment_idx ON posts (sentiment);
CREATE INDEX reports_created_at_idx ON reports (created_at);
CREATE INDEX reports_reported_by_idx ON reports (reported_by);
CREATE INDEX saved_comments_user_id_idx ON saved_comments (user_id);
CREATE INDEX saved_comments_comment_id_idx ON saved_comments (comment_id);
CREATE INDEX saved_posts_post_id_idx ON saved_posts (post_id);
CREATE INDEX saved_posts_user_id_idx ON saved_posts (user_id);
CREATE INDEX school_follows_user_id_idx ON school_follows (user_id);
CREATE INDEX votes_post_id_idx ON votes (post_id);
CREATE INDEX votes_comment_id_idx ON votes (comment_id);


END;
