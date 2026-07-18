-- name: CreateUser :one
INSERT INTO users (email, password_hash, created_at, updated_at, is_active, mfa_enabled, mfa_secret)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, email, password_hash, created_at, updated_at, is_active, mfa_enabled, mfa_secret;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at, updated_at, is_active, mfa_enabled, mfa_secret
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, created_at, updated_at, is_active, mfa_enabled, mfa_secret
FROM users 
WHERE id = $1;

-- name: EnableMFASecret :exec
UPDATE users SET mfa_secret = $2, updated_at = NOW() WHERE id = $1;

-- name: ConfirmMFA :exec
UPDATE users SET mfa_enabled = TRUE, updated_at = NOW() WHERE id = $1;

-- name: DisableMFA :exec
UPDATE users SET mfa_enabled = FALSE, mfa_secret = NULL, updated_at = NOW() WHERE id = $1;

-- name: UpdateUserPassword :execrows
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;