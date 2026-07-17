package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"vault_api/internal/api/middleware"
	"vault_api/internal/domain"
	"vault_api/internal/service"
)

func (h *Handler) ShareItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	itemID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid item id", http.StatusBadRequest)
		return
	}

	var req struct {
		Email            string `json:"email"`
		EncryptedItemKey []byte `json:"encrypted_item_key"`
		Permission       string `json:"permission"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Email) == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Permission) == "" {
		http.Error(w, "permission is required", http.StatusBadRequest)
		return
	}

	share, err := h.vaultService.ShareItem(r.Context(), userID, itemID, auditContextFromRequest(r), service.ShareVaultItemInput{
		Email:            req.Email,
		EncryptedItemKey: req.EncryptedItemKey,
		Permission:       req.Permission,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEncryptedItemKey):
			http.Error(w, "invalid encrypted item key", http.StatusBadRequest)
		case errors.Is(err, service.ErrInvalidSharePermission):
			http.Error(w, "permission must be read or write", http.StatusBadRequest)
		case errors.Is(err, service.ErrCannotShareWithSelf):
			http.Error(w, "cannot share item with yourself", http.StatusBadRequest)
		case errors.Is(err, service.ErrAlreadyShared):
			http.Error(w, "item already shared with user", http.StatusConflict)
		case errors.Is(err, service.ErrNotFound):
			http.Error(w, "item or recipient not found", http.StatusNotFound)
		default:
			http.Error(w, "failed to share item", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusCreated, share)
}

func (h *Handler) RevokeShare(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	itemID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid item id", http.StatusBadRequest)
		return
	}

	recipientID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.vaultService.RevokeShare(r.Context(), userID, itemID, recipientID, auditContextFromRequest(r)); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "share not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to revoke share", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListSharedItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	filter, err := parseVaultListFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.vaultService.ListSharedWithMe(r.Context(), userID, filter.Limit, filter.Offset)
	if err != nil {
		http.Error(w, "failed to list shared items", http.StatusInternalServerError)
		return
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, entry := range result.Items {
		items = append(items, sharedVaultItemResponse(entry))
	}
	if len(items) == 0 {
		items = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"total":  result.Total,
		"limit":  result.Limit,
		"offset": result.Offset,
	})
}

func sharedVaultItemResponse(entry domain.SharedVaultItemWithItem) map[string]any {
	return map[string]any{
		"id":                 entry.Share.ID,
		"vault_item_id":      entry.Share.VaultItemID,
		"owner_id":           entry.Share.OwnerID,
		"shared_with_user_id": entry.Share.SharedWithUserID,
		"encrypted_item_key": entry.Share.EncryptedItemKey,
		"permission":         entry.Share.Permission,
		"created_at":         entry.Share.CreatedAt,
		"item":               entry.Item,
	}
}
