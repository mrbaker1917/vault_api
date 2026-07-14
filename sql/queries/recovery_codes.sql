-- name: CreateRecoveryCode :one
INSERT INTO recovery_codes (user_id, code_hash, created_at)
VALUES ($1, $2, NOW())
RETURNING id, user_id, code_hash, used_at, created_at;

-- name: DeleteRecoveryCodesByUserID :exec
DELETE FROM recovery_codes WHERE user_id = $1;

-- name: ListUnusedRecoveryCodesByUserID :many
SELECT id, user_id, code_hash, used_at, created_at
FROM recovery_codes
WHERE user_id = $1
  AND used_at IS NULL;

-- name: MarkRecoveryCodeUsed :execrows
UPDATE recovery_codes
SET used_at = NOW()
WHERE id = $1
  AND used_at IS NULL;
