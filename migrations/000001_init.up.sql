BEGIN;
    CREATE TABLE mod_levels (
        id INTEGER PRIMARY KEY UNIQUE NOT NULL,
        mod VARCHAR(20) UNIQUE NOT NULL
    );

    INSERT INTO mod_levels (id, mod) VALUES 
        (1, 'enabled'),
        (2, 'banned'),
        (3, 'limited');

    CREATE TABLE schools (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        name VARCHAR(255) NOT NULL,
        abbr VARCHAR(10) NOT NULL,
        lat FLOAT4 NOT NULL,
        lon FLOAT4 NOT NULL,
        domain VARCHAR(100) NOT NULL UNIQUE,
        CONSTRAINT sch_uniq_coord UNIQUE(lat, lon)
    );


    CREATE TABLE faculties (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        faculty VARCHAR(255) UNIQUE NOT NULL
    );

    CREATE TABLE users (
        id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
        created_at TIMESTAMPTZ,
        updated_at TIMESTAMPTZ,
        email VARCHAR(100) NOT NULL,
        year_of_study INTEGER,
        faculty_id INTEGER,
        school_id INTEGER,
        mod_id INTEGER,
        CONSTRAINT us_fk_sch FOREIGN KEY (school_id) REFERENCES schools (id),
        CONSTRAINT us_fk_fac FOREIGN KEY (faculty_id) REFERENCES faculties (id)
    );

    CREATE TABLE school_follows (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        school_id INTEGER NOT NULL,
        CONSTRAINT sf_uniq UNIQUE(user_id, school_id),
        CONSTRAINT sf_fk_use 
            FOREIGN KEY (user_id)
            REFERENCES users (id) 
            ON DELETE CASCADE,
        CONSTRAINT sf_fk_sch FOREIGN KEY (school_id) REFERENCES schools (id)
    );


    CREATE TABLE posts (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ,
        user_id VARCHAR(100) NOT NULL,
        school_id INTEGER NOT NULL,
        faculty_id INTEGER,
        title VARCHAR(255),
        content TEXT,
        downvote INTEGER,
        upvote INTEGER,
        vote_score INTEGER,
        trending_score FLOAT4,
        hottest_on DATE,
        hidden BOOLEAN,
        CONSTRAINT po_fk_use FOREIGN KEY (user_id) REFERENCES users (id),
        CONSTRAINT po_fk_sch FOREIGN KEY (school_id) REFERENCES schools (id),
        CONSTRAINT po_fk_fal FOREIGN KEY (faculty_id) REFERENCES faculties (id)
    );

    CREATE TABLE comments (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ,
        user_id VARCHAR(255) NOT NULL,
        post_id INTEGER NOT NULL,
        comment_id INTEGER,
        content TEXT NOT NULL,
        upvote INTEGER,
        downvote INTEGER,
        score INTEGER,
        hidden BOOLEAN,
        CONSTRAINT co_fk_use FOREIGN KEY (user_id) REFERENCES users (id),
        CONSTRAINT co_fk_pos 
            FOREIGN KEY (post_id) 
            REFERENCES posts (id) 
            ON DELETE CASCADE
    );

    CREATE TYPE vote_score_value AS ENUM ('1','-1');

    CREATE TABLE votes (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        vote vote_score_value NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        post_id INTEGER NOT NULL,
        comment_id INTEGER,
        CONSTRAINT vot_fk_use FOREIGN KEY (user_id) REFERENCES users (id),
        CONSTRAINT vot_fk_pos FOREIGN KEY (post_id) REFERENCES posts (id),
        CONSTRAINT vot_fk_com FOREIGN KEY (comment_id) REFERENCES comments (id),
        CONSTRAINT vot_uniq_vote UNIQUE (user_id, post_id, comment_id)
    );

    CREATE TABLE saved_posts (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        created_at TIMESTAMPTZ NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        post_id INTEGER NOT NULL,
        CONSTRAINT sap_fk_use FOREIGN KEY (user_id) REFERENCES users (id),
        CONSTRAINT sap_fk_pos FOREIGN KEY (post_id) REFERENCES posts (id)
    );

    CREATE TABLE saved_comments (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        created_at TIMESTAMPTZ NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        comment_id INTEGER NOT NULL,
        CONSTRAINT sco_fk_use 
            FOREIGN KEY (user_id) 
            REFERENCES users (id) 
            ON DELETE CASCADE,
        CONSTRAINT sco_fk_pos 
            FOREIGN KEY (comment_id) 
            REFERENCES comments (id) 
            ON DELETE CASCADE
    );

    CREATE TABLE feedbacks (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        created_at TIMESTAMPTZ NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        content TEXT NOT NULL,
        CONSTRAINT feb_fk_use 
            FOREIGN KEY (user_id) 
            REFERENCES users (id) 
            ON DELETE CASCADE
    );

    CREATE TABLE reports (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        created_at TIMESTAMPTZ NOT NULL,
        reported_by VARCHAR(255) NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        description TEXT NOT NULL,
        type VARCHAR(255) NOT NULL,
        result TEXT,
        user_alerted BOOLEAN DEFAULT false,
        CONSTRAINT rep_fk_use FOREIGN KEY (user_id) REFERENCES users (id),
        CONSTRAINT rep_fk_rep FOREIGN KEY (reported_by) REFERENCES users (id)
    );

END;
