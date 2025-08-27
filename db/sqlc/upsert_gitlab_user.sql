-- UpsertGitlabUser upserts a gitlab user
-- name: UpsertGitlabUser :exec
INSERT INTO gitlab_user_cache (gitlab_user_id, email, updated_at)
    VALUES (@gitlab_user_id, @email, NOW())
ON CONFLICT (gitlab_user_id)
    DO UPDATE SET
        email = EXCLUDED.email,
        updated_at = EXCLUDED.updated_at;

