-- name: CreateVaultItem :one
INSERT INTO vault_items (user_id, encrypted_data, item_type, title, folder, tags, created_at, updated_at, version)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, encrypted_data, item_type, title, folder, tags, created_at, updated_at, deleted_at, version;

-- name: GetVaultItemByID :one
SELECT id, user_id, encrypted_data, item_type, title, folder, tags, created_at, updated_at, deleted_at, version
FROM vault_items
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListVaultItemsByUserID :many
SELECT id, user_id, encrypted_data, item_type, title, folder, tags, created_at, updated_at, deleted_at, version
FROM vault_items
WHERE user_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC;
