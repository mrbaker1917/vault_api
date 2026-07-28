-- name: CreateAuditLog :one
INSERT INTO audit_log (user_id, action, resource_type, resource_id, ip_address, user_agent, metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
RETURNING id, user_id, action, resource_type, resource_id, ip_address, user_agent, metadata, created_at;

-- name: ListAuditLogsByUserID :many
SELECT id, user_id, action, resource_type, resource_id, ip_address, user_agent, metadata, created_at
FROM audit_log
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsFiltered :many
SELECT id, user_id, action, resource_type, resource_id, ip_address, user_agent, metadata, created_at
FROM audit_log
WHERE user_id = $1
  AND (NULLIF($2, '') IS NULL OR action LIKE $2 || '%')
  AND (NULLIF($3, '') IS NULL OR action = $3)
  AND ($4::timestamptz IS NULL OR created_at >= $4)
  AND ($5::timestamptz IS NULL OR created_at <= $5)
ORDER BY created_at DESC
LIMIT $6 OFFSET $7;
