package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"vault_api/internal/domain"
	"vault_api/internal/repository/sqlc"
)

type vaultItemPostgresRepository struct {
	q *sqlc.Queries
}

func NewVaultItemRepository(pg *Postgres) VaultItemRepository {
	return &vaultItemPostgresRepository{
		q: sqlc.New(pg.Pool()),
	}
}

func (r *vaultItemPostgresRepository) Create(ctx context.Context, item domain.VaultItem) (domain.VaultItem, error) {
	row, err := r.q.CreateVaultItem(ctx, sqlc.CreateVaultItemParams{
		UserID:        pgUUIDToPG(item.UserID),
		EncryptedData: item.EncryptedData,
		ItemType:      item.ItemType,
		Title:         pgTextFromString(item.Title),
		Folder:        pgTextFromString(item.Folder),
		Tags:          item.Tags,
		CreatedAt:     pgTimestampToPG(item.CreatedAt),
		UpdatedAt:     pgTimestampToPG(item.UpdatedAt),
		Version:       pgInt4ToPG(item.Version),
	})
	if err != nil {
		return domain.VaultItem{}, fmt.Errorf("create vault item: %w", err)
	}
	vaultItemRow, err := vaultItemRowFromSQLC(row)
	if err != nil {
		return domain.VaultItem{}, fmt.Errorf("create vault item: %w", err)
	}
	return toDomainVaultItem(vaultItemRow), nil
}

func (r *vaultItemPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.VaultItem, error) {
	row, err := r.q.GetVaultItemByID(ctx, pgUUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VaultItem{}, ErrNotFound
		}
		return domain.VaultItem{}, fmt.Errorf("get vault item by id: %w", err)
	}
	vaultItemRow, err := vaultItemRowFromSQLC(row)
	if err != nil {
		return domain.VaultItem{}, fmt.Errorf("get vault item by id: %w", err)
	}
	return toDomainVaultItem(vaultItemRow), nil
}

func (r *vaultItemPostgresRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.VaultItem, error) {
	rows, err := r.q.ListVaultItemsByUserID(ctx, pgUUIDToPG(userID))
	if err != nil {
		return nil, fmt.Errorf("list vault items by user id: %w", err)
	}
	items := make([]domain.VaultItem, 0, len(rows))
	for _, row := range rows {
		vaultItemRow, err := vaultItemRowFromSQLC(row)
		if err != nil {
			return nil, fmt.Errorf("list vault items by user id: %w", err)
		}
		items = append(items, toDomainVaultItem(vaultItemRow))
	}
	return items, nil
}

type vaultItemRow struct {
	id            uuid.UUID
	userID        uuid.UUID
	encryptedData []byte
	itemType      string
	title         string
	folder        string
	tags          []string
	createdAt     time.Time
	updatedAt     time.Time
	deletedAt     *time.Time
	version       int32
}

func vaultItemRowFromSQLC(r sqlc.VaultItem) (vaultItemRow, error) {
	id, err := uuidFromPG(r.ID)
	if err != nil {
		return vaultItemRow{}, err
	}
	userID, err := uuidFromPG(r.UserID)
	if err != nil {
		return vaultItemRow{}, err
	}
	return vaultItemRow{
		id:            id,
		userID:        userID,
		encryptedData: r.EncryptedData,
		itemType:      r.ItemType,
		title:         pgTextToString(r.Title),
		folder:        pgTextToString(r.Folder),
		tags:          r.Tags,
		createdAt:     pgTimestampFromPG(r.CreatedAt),
		updatedAt:     pgTimestampFromPG(r.UpdatedAt),
		deletedAt:     pgTimestampToPtr(r.DeletedAt),
		version:       pgInt4FromPG(r.Version),
	}, nil
}

func toDomainVaultItem(row vaultItemRow) domain.VaultItem {
	return domain.VaultItem{
		ID:            row.id,
		UserID:        row.userID,
		EncryptedData: row.encryptedData,
		ItemType:      row.itemType,
		Title:         row.title,
		Folder:        row.folder,
		Tags:          row.tags,
		CreatedAt:     row.createdAt,
		UpdatedAt:     row.updatedAt,
		DeletedAt:     row.deletedAt,
		Version:       row.version,
	}
}

func (r *vaultItemPostgresRepository) Update(ctx context.Context, item domain.VaultItem) (domain.VaultItem, error) {
	row, err := r.q.UpdateVaultItem(ctx, sqlc.UpdateVaultItemParams{
		ID:            pgUUIDToPG(item.ID),
		EncryptedData: item.EncryptedData,
		ItemType:      item.ItemType,
		Title:         pgTextFromString(item.Title),
		Folder:        pgTextFromString(item.Folder),
		Tags:          item.Tags,
		UpdatedAt:     pgTimestampToPG(item.UpdatedAt),
		Version:       pgInt4ToPG(item.Version),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VaultItem{}, ErrNotFound
		}
		return domain.VaultItem{}, fmt.Errorf("update vault item: %w", err)
	}
	vaultItemRow, err := vaultItemRowFromSQLC(row)
	if err != nil {
		return domain.VaultItem{}, fmt.Errorf("update vault item: %w", err)
	}
	return toDomainVaultItem(vaultItemRow), nil
}

func (r *vaultItemPostgresRepository) Delete(ctx context.Context, id uuid.UUID, version int32) (domain.VaultItem, error) {
	row, err := r.q.DeleteVaultItem(ctx, sqlc.DeleteVaultItemParams{
		ID:      pgUUIDToPG(id),
		Version: pgInt4ToPG(version),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VaultItem{}, ErrNotFound
		}
		return domain.VaultItem{}, fmt.Errorf("delete vault item: %w", err)
	}
	vaultItemRow, err := vaultItemRowFromSQLC(row)
	if err != nil {
		return domain.VaultItem{}, fmt.Errorf("delete vault item: %w", err)
	}
	return toDomainVaultItem(vaultItemRow), nil
}

func (r *vaultItemPostgresRepository) Restore(ctx context.Context, id uuid.UUID, version int32, userID uuid.UUID) (domain.VaultItem, error) {
	row, err := r.q.RestoreVaultItem(ctx, sqlc.RestoreVaultItemParams{
		ID:      pgUUIDToPG(id),
		Version: pgInt4ToPG(version),
		UserID:  pgUUIDToPG(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VaultItem{}, ErrNotFound
		}
		return domain.VaultItem{}, fmt.Errorf("restore vault item: %w", err)
	}
	vaultItemRow, err := vaultItemRowFromSQLC(row)
	if err != nil {
		return domain.VaultItem{}, fmt.Errorf("restore vault item: %w", err)
	}
	return toDomainVaultItem(vaultItemRow), nil
}