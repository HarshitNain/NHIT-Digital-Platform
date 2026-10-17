-- name: CreateSession :one
INSERT INTO sessions (
    session_id, user_id, session_token, created_at, expires_at
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions
WHERE session_id = $1;

-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE session_token = $1;

-- name: GetUserSessions :many
SELECT * FROM sessions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE session_id = $1;

-- name: DeleteUserSessions :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < NOW();

-- name: CountUserSessions :one
SELECT COUNT(*) FROM sessions
WHERE user_id = $1;
