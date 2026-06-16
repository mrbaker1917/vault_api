-- name: CreateUser :one
INSERT INTO users (email, password_hash, created_at, updated_at, is_active, mfa_enabled, mfa_secret)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, email, password_hash, created_at, updated_at, is_active, mfa_enabled, mfa_secret;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at, updated_at, is_active, mfa_enabled, mfa_secret
FROM users
WHERE email = $1;