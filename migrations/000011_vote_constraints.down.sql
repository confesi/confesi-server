BEGIN;

ALTER TABLE votes
    ALTER COLUMN post_id SET NOT NULL;

    ALTER TYPE vote_score_value DROP VALUE '0';
END;
