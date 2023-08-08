BEGIN;    

CREATE TYPE role_action_type AS  ENUM ('set','add', 'remove');

    CREATE TABLE role_assignment_logs (
        id SERIAL PRIMARY KEY UNIQUE NOT NULL,
        created_at TIMESTAMPTZ NOT NULL,
        action_user_id VARCHAR(255) NOT NULL,
        affected_user_id VARCHAR(255) NOT NULL,
        old_roles VARCHAR(255)[] NOT NULL,
        new_roles VARCHAR(255)[] NOT NULL,
        action_type role_action_type NOT NULL
        
    );
END;