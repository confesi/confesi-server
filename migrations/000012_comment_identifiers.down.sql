BEGIN;
    ALTER TABLE comments
        DROP COLUMN identifier_id;

    DROP TABLE IF EXISTS comment_identifiers;
END;
