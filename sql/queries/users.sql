-- name: GetUserByGoogleID :one
SELECT id, email, name, avatar_url, phone, google_id, created_at, updated_at
FROM users
WHERE google_id = $1;

-- name: CreateUser :one
INSERT INTO users (email, name, avatar_url, phone, google_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, name, avatar_url, phone, google_id, created_at, updated_at;

-- name: UpdateUserPhoneByGoogleID :exec
UPDATE users
SET phone = $2,
    updated_at = now()
WHERE google_id = $1
  AND $2::text <> '';
