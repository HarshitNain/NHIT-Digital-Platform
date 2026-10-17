-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    token, user_id, expires_at, created_at
) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: GetUserRefreshTokens :many
SELECT * FROM refresh_tokens
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = $1;

-- name: DeleteUserRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE expires_at < NOW();

-- name: UpdateRefreshTokenExpiry :exec
UPDATE refresh_tokens
SET expires_at = $2
WHERE token = $1;
