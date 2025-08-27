-- UpsertContractDirectory upserts a contract directory entry
-- name: UpsertContractDirectory :exec
INSERT INTO contract (namespace, queue, name, payload, updated_at)
    VALUES (@namespace, @queue, @name, @payload, NOW())
ON CONFLICT (@namespace, @queue, @name)
    DO UPDATE SET
        @payload = EXCLUDED.payload,
        updated_at = EXCLUDED.updated_at;
