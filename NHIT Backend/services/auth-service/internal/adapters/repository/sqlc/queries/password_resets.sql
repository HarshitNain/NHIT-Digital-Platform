-- name: CreatePasswordReset :one
INSERT INTO password_resets (
    token, user_id, expires_at, created_at
) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPasswordReset :one
SELECT * FROM password_resets
WHERE token = $1;

-- name: DeletePasswordReset :exec
DELETE FROM password_resets
WHERE token = $1;

-- name: DeleteUserPasswordResets :exec
DELETE FROM password_resets
WHERE user_id = $1;

-- name: DeleteExpiredPasswordResets :exec
DELETE FROM password_resets
WHERE expires_at < NOW();
