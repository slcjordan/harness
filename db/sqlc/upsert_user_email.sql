-- UpsertUserEmail upserts a user email
-- name: UpsertUserEmail :one
INSERT INTO user_email (user_id, address, status, updated_at)
    VALUES (@user_id, @address, @status, NOW())
ON CONFLICT (gitlab_user_id)
    DO UPDATE SET
        user_id = EXCLUDED.user_id,
        address = EXCLUDED.address,
        status = EXCLUDED.status,
        updated_at = NOW()
    RETURNING
        id;

