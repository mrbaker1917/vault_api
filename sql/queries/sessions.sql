-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, device_name, ip_address, user_agent, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, token_hash, device_name, ip_address, user_agent, created_at, expires_at, revoked_at;

-- name: GetSessionByTokenHash :one
SELECT id, user_id, token_hash, device_name, ip_address, user_agent, created_at, expires_at, revoked_at
FROM sessions
WHERE token_hash = $1;

-- name: GetSessionByID :one
SELECT id, user_id, token_hash, device_name, ip_address, user_agent,
       created_at, expires_at, revoked_at
FROM sessions
WHERE id = $1
  AND revoked_at IS NULL
  AND expires_at > NOW();