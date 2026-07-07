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

-- name: UpdateVaultItem :one
UPDATE vault_items
SET encrypted_data = $2, item_type = $3, title = $4, folder = $5, tags = $6, updated_at = $7, version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $8
RETURNING id, user_id, encrypted_data, item_type, title, folder, tags, created_at, updated_at, deleted_at, version;

-- name: DeleteVaultItem :one
UPDATE vault_items  
SET deleted_at = NOW(), version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $2
RETURNING id, user_id, encrypted_data, item_type, title, folder, tags, created_at, updated_at, deleted_at, version;

-- name: RestoreVaultItem :one
UPDATE vault_items
SET deleted_at = NULL, version = version + 1
WHERE id = $1
  AND deleted_at IS NOT NULL
  AND version = $2
  AND user_id = $3
RETURNING id, user_id, encrypted_data, item_type, title, folder, tags, created_at, updated_at, deleted_at, version;