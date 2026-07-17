-- name: CreateSharedVaultItem :one
INSERT INTO shared_vault_items (vault_item_id, owner_id, shared_with_user_id, encrypted_item_key, permission, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, vault_item_id, owner_id, shared_with_user_id, encrypted_item_key, permission, created_at;

-- name: GetShareByVaultAndRecipient :one
SELECT id, vault_item_id, owner_id, shared_with_user_id, encrypted_item_key, permission, created_at
FROM shared_vault_items
WHERE vault_item_id = $1
  AND shared_with_user_id = $2;

-- name: DeleteSharedVaultItem :execrows
DELETE FROM shared_vault_items
WHERE vault_item_id = $1
  AND owner_id = $2
  AND shared_with_user_id = $3;

-- name: ListSharedVaultItemsForUser :many
SELECT
  s.id,
  s.vault_item_id,
  s.owner_id,
  s.shared_with_user_id,
  s.encrypted_item_key,
  s.permission,
  s.created_at,
  v.id AS item_id,
  v.user_id AS item_user_id,
  v.encrypted_data AS item_encrypted_data,
  v.item_type AS item_item_type,
  v.title AS item_title,
  v.folder AS item_folder,
  v.tags AS item_tags,
  v.created_at AS item_created_at,
  v.updated_at AS item_updated_at,
  v.deleted_at AS item_deleted_at,
  v.version AS item_version
FROM shared_vault_items s
INNER JOIN vault_items v ON v.id = s.vault_item_id
WHERE s.shared_with_user_id = $1
  AND v.deleted_at IS NULL
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSharedVaultItemsForUser :one
SELECT COUNT(*)
FROM shared_vault_items s
INNER JOIN vault_items v ON v.id = s.vault_item_id
WHERE s.shared_with_user_id = $1
  AND v.deleted_at IS NULL;
