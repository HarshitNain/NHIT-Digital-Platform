-- name: CreateEmailVerificationToken :one
INSERT INTO email_verification_tokens (
    token, user_id, expires_at, created_at
) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetEmailVerificationToken :one
SELECT * FROM email_verification_tokens
WHERE token = $1;

-- name: GetEmailVerificationByUser :one
SELECT * FROM email_verification_tokens
WHERE user_id = $1;

-- name: DeleteEmailVerificationToken :exec
DELETE FROM email_verification_tokens
WHERE token = $1;

-- name: DeleteUserEmailVerificationToken :exec
DELETE FROM email_verification_tokens
WHERE user_id = $1;

-- name: DeleteExpiredEmailVerificationTokens :exec
DELETE FROM email_verification_tokens
WHERE expires_at < NOW();
