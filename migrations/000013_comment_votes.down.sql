BEGIN;

ALTER TABLE comments
    ADD COLUMN score INTEGER ,
    DROP COLUMN trending_score,
    DROP COLUMN vote_score;

END;
