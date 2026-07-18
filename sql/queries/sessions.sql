-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, device_name, ip_address, user_agent, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, token_hash, device_name, ip_address, user_agent, created_at, expires_at, revoked_at;

-- name: GetSessionByTokenHash :one
SELECT id, user_id, token_hash, device_name, ip_address, user_agent, created_at, expires_at, revoked_at
FROM sessions
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND expires_at > NOW();

-- name: GetSessionByID :one
SELECT id, user_id, token_hash, device_name, ip_address, user_agent,
       created_at, expires_at, revoked_at
FROM sessions
WHERE id = $1
  AND revoked_at IS NULL
  AND expires_at > NOW();

-- name: RevokeSession :execrows
UPDATE sessions
SET revoked_at = NOW()
WHERE id = $1
  AND revoked_at IS NULL
  AND user_id = $2;

-- name: ListSessionsByUserID :many
SELECT id, user_id, device_name, ip_address, user_agent, created_at, expires_at, revoked_at
FROM sessions
WHERE user_id = $1
  AND revoked_at IS NULL
  AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: RevokeSessionsExcept :exec
UPDATE sessions
SET revoked_at = NOW()
WHERE user_id = $1
  AND id <> $2
  AND revoked_at IS NULL;
