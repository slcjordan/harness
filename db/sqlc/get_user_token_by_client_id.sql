-- GetUserTokenByClientID fetches the most-recent token for a user id.
-- name: GetUserTokenByClientID :one
SELECT
    authed_user_access_token
FROM
    slack_oauth_response
WHERE
    authed_user_id = @authed_user_id
ORDER BY
    created_at DESC
LIMIT 1;

