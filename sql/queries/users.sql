-- name: GetUserByGoogleID :one
SELECT id, email, name, avatar_url, google_id, created_at, updated_at
FROM users
WHERE google_id = $1;

-- name: CreateUser :one
INSERT INTO users (email, name, avatar_url, google_id)
VALUES ($1, $2, $3, $4)
RETURNING id, email, name, avatar_url, google_id, created_at, updated_at;
