-- Remove Indexs from columns that are used in WHERE clauses if they exist
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

DROP INDEX IF EXISTS comments_hidden_idx;
DROP INDEX IF EXISTS comments_user_id_idx;
DROP INDEX IF EXISTS comments_post_id_idx;
DROP INDEX IF EXISTS comments_created_at_idx;
DROP INDEX IF EXISTS comments_trending_score_idx;
DROP INDEX IF EXISTS users_school_id_idx;
DROP INDEX IF EXISTS drafts_user_id_idx;
DROP INDEX IF EXISTS fcm_tokens_user_id_idx;
DROP INDEX IF EXISTS fcm_topic_prefs_user_id_idx;
DROP INDEX IF EXISTS posts_school_id_idx;
DROP INDEX IF EXISTS posts_user_id_idx;
DROP INDEX IF EXISTS posts_sentiment_idx;
DROP INDEX IF EXISTS reports_created_at_idx;
DROP INDEX IF EXISTS reports_reported_by_idx;
DROP INDEX IF EXISTS saved_comments_user_id_idx;
DROP INDEX IF EXISTS saved_comments_comment_id_idx;
DROP INDEX IF EXISTS saved_posts_post_id_idx;
DROP INDEX IF EXISTS saved_posts_user_id_idx;
DROP INDEX IF EXISTS school_follows_user_id_idx;
DROP INDEX IF EXISTS votes_post_id_idx;
DROP INDEX IF EXISTS votes_comment_id_idx;

END;
