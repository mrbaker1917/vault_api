package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"vault_api/internal/domain"
	"vault_api/internal/repository/sqlc"
)

type sharedVaultItemPostgresRepository struct {
	q *sqlc.Queries
}

func NewSharedVaultItemRepository(pg *Postgres) SharedVaultItemRepository {
	return &sharedVaultItemPostgresRepository{
		q: sqlc.New(pg.Pool()),
	}
}

func (r *sharedVaultItemPostgresRepository) Create(ctx context.Context, share domain.SharedVaultItem) (domain.SharedVaultItem, error) {
	row, err := r.q.CreateSharedVaultItem(ctx, sqlc.CreateSharedVaultItemParams{
		VaultItemID:      pgUUIDToPG(share.VaultItemID),
		OwnerID:          pgUUIDToPG(share.OwnerID),
		SharedWithUserID: pgUUIDToPG(share.SharedWithUserID),
		EncryptedItemKey: share.EncryptedItemKey,
		Permission:       share.Permission,
		CreatedAt:        pgTimestampToPG(share.CreatedAt),
	})
	if err != nil {
		return domain.SharedVaultItem{}, fmt.Errorf("create shared vault item: %w", err)
	}
	return toDomainSharedVaultItem(row)
}

func (r *sharedVaultItemPostgresRepository) GetByVaultAndRecipient(ctx context.Context, vaultItemID, recipientID uuid.UUID) (domain.SharedVaultItem, error) {
	row, err := r.q.GetShareByVaultAndRecipient(ctx, sqlc.GetShareByVaultAndRecipientParams{
		VaultItemID:      pgUUIDToPG(vaultItemID),
		SharedWithUserID: pgUUIDToPG(recipientID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SharedVaultItem{}, ErrNotFound
		}
		return domain.SharedVaultItem{}, fmt.Errorf("get shared vault item: %w", err)
	}
	return toDomainSharedVaultItem(row)
}

func (r *sharedVaultItemPostgresRepository) Delete(ctx context.Context, vaultItemID, ownerID, recipientID uuid.UUID) error {
	rowsAffected, err := r.q.DeleteSharedVaultItem(ctx, sqlc.DeleteSharedVaultItemParams{
		VaultItemID:      pgUUIDToPG(vaultItemID),
		OwnerID:          pgUUIDToPG(ownerID),
		SharedWithUserID: pgUUIDToPG(recipientID),
	})
	if err != nil {
		return fmt.Errorf("delete shared vault item: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sharedVaultItemPostgresRepository) ListForRecipient(ctx context.Context, recipientID uuid.UUID, limit, offset int32) (ListSharedVaultItemsResult, error) {
	pgRecipientID := pgUUIDToPG(recipientID)

	total, err := r.q.CountSharedVaultItemsForUser(ctx, pgRecipientID)
	if err != nil {
		return ListSharedVaultItemsResult{}, fmt.Errorf("count shared vault items: %w", err)
	}

	rows, err := r.q.ListSharedVaultItemsForUser(ctx, sqlc.ListSharedVaultItemsForUserParams{
		SharedWithUserID: pgRecipientID,
		Limit:            limit,
		Offset:           offset,
	})
	if err != nil {
		return ListSharedVaultItemsResult{}, fmt.Errorf("list shared vault items: %w", err)
	}

	items := make([]domain.SharedVaultItemWithItem, 0, len(rows))
	for _, row := range rows {
		entry, err := toDomainSharedVaultItemWithItem(row)
		if err != nil {
			return ListSharedVaultItemsResult{}, fmt.Errorf("list shared vault items: %w", err)
		}
		items = append(items, entry)
	}

	return ListSharedVaultItemsResult{
		Items: items,
		Total: total,
	}, nil
}

func toDomainSharedVaultItem(row sqlc.SharedVaultItem) (domain.SharedVaultItem, error) {
	id, err := uuidFromPG(row.ID)
	if err != nil {
		return domain.SharedVaultItem{}, err
	}
	vaultItemID, err := uuidFromPG(row.VaultItemID)
	if err != nil {
		return domain.SharedVaultItem{}, err
	}
	ownerID, err := uuidFromPG(row.OwnerID)
	if err != nil {
		return domain.SharedVaultItem{}, err
	}
	sharedWithUserID, err := uuidFromPG(row.SharedWithUserID)
	if err != nil {
		return domain.SharedVaultItem{}, err
	}
	return domain.SharedVaultItem{
		ID:               id,
		VaultItemID:      vaultItemID,
		OwnerID:          ownerID,
		SharedWithUserID: sharedWithUserID,
		EncryptedItemKey: row.EncryptedItemKey,
		Permission:       row.Permission,
		CreatedAt:        pgTimestampFromPG(row.CreatedAt),
	}, nil
}

func toDomainSharedVaultItemWithItem(row sqlc.ListSharedVaultItemsForUserRow) (domain.SharedVaultItemWithItem, error) {
	share, err := toDomainSharedVaultItem(sqlc.SharedVaultItem{
		ID:               row.ID,
		VaultItemID:      row.VaultItemID,
		OwnerID:          row.OwnerID,
		SharedWithUserID: row.SharedWithUserID,
		EncryptedItemKey: row.EncryptedItemKey,
		Permission:       row.Permission,
		CreatedAt:        row.CreatedAt,
	})
	if err != nil {
		return domain.SharedVaultItemWithItem{}, err
	}

	itemID, err := uuidFromPG(row.ItemID)
	if err != nil {
		return domain.SharedVaultItemWithItem{}, err
	}
	itemUserID, err := uuidFromPG(row.ItemUserID)
	if err != nil {
		return domain.SharedVaultItemWithItem{}, err
	}

	return domain.SharedVaultItemWithItem{
		Share: share,
		Item: domain.VaultItem{
			ID:            itemID,
			UserID:        itemUserID,
			EncryptedData: row.ItemEncryptedData,
			ItemType:      row.ItemItemType,
			Title:         pgTextToString(row.ItemTitle),
			Folder:        pgTextToString(row.ItemFolder),
			Tags:          row.ItemTags,
			CreatedAt:     pgTimestampFromPG(row.ItemCreatedAt),
			UpdatedAt:     pgTimestampFromPG(row.ItemUpdatedAt),
			DeletedAt:     pgTimestampToPtr(row.ItemDeletedAt),
			Version:       pgInt4FromPG(row.ItemVersion),
		},
	}, nil
}
