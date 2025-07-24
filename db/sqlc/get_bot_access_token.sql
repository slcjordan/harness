-- GetBotAccessToken fetches the most-recent token for a user id.
-- name: GetBotAccessToken :one
SELECT
    access_token
FROM
    slack_oauth_response
ORDER BY
    created_at DESC
LIMIT 1;

