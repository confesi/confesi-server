BEGIN;
    DROP TABLE IF EXISTS
        mod_levels,
        schools,
        faculties,
        users,
        school_follows,
        posts,
        comments,
        votes,
        saved_posts,
        saved_comments,
        feedbacks,
        reports
    ;

    DROP TYPE IF EXISTS vote_score_value;
END;
