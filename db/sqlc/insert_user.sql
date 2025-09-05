-- InsertUser inserts a new user
-- name: InsertUser :one
INSERT INTO "user" DEFAULT
    VALUES
    RETURNING
        id;

