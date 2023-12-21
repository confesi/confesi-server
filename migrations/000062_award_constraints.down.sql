BEGIN;

-- Re-add the CHECK constraint
ALTER TABLE awards_general ADD CONSTRAINT awards_general_post_id_comment_id_check CHECK (
    (post_id IS NOT NULL AND comment_id IS NULL) OR 
    (post_id IS NULL AND comment_id IS NOT NULL)
);

END;
