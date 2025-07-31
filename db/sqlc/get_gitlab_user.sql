-- GetGitlabUser retrieves a gitlab user from the user cache.
-- name: GetGitlabUser :one
SELECT
    email
FROM
    gitlab_user_cache
WHERE
    gitlab_user_id = @gitlab_user_id
    AND updated_at >= NOW() - INTERVAL '60 days';

