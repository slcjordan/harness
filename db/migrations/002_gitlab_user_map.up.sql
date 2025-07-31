CREATE TABLE gitlab_user_cache (
    gitlab_user_id integer PRIMARY KEY,
    email text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

