
BEGIN;

ALTER TABLE posts ADD COLUMN sentiment_score FLOAT NOT NULL DEFAULT 0.0; 

END;
