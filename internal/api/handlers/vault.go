package handlers

import (
	"net/http"
	"errors"
	"strings"
	"vault_api/internal/api/middleware"
	"vault_api/internal/service"
	"github.com/google/uuid"
	"encoding/json"
	"vault_api/internal/domain"
)

func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		EncryptedData []byte   `json:"encrypted_data"` // base64 in JSON, or string
		ItemType      string   `json:"item_type"`
		Title         string   `json:"title"`
		Folder        string   `json:"folder"`
		Tags          []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.ItemType) == "" {
		http.Error(w, "item_type is required", http.StatusBadRequest)
		return
	}

	input := service.CreateVaultItemInput{
		EncryptedData: req.EncryptedData,
		ItemType:      req.ItemType,
		Title:         req.Title,
		Folder:        req.Folder,
		Tags:          req.Tags,
	}

	item, err := h.vaultService.CreateItem(r.Context(), userID, auditContextFromRequest(r), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEncryptedBlob) {
			http.Error(w, "invalid encrypted blob", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to create item", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.vaultService.ListItems(r.Context(), userID, filter)
	if err != nil {
		http.Error(w, "failed to list items", http.StatusInternalServerError)
		return
	}

	if len(result.Items) == 0 {
		result.Items = []domain.VaultItem{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  result.Items,
		"total":  result.Total,
		"limit":  result.Limit,
		"offset": result.Offset,
	})
}

func (h *Handler) ListDeletedItems(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.vaultService.ListDeletedItems(r.Context(), userID, filter.Limit, filter.Offset)
	if err != nil {
		http.Error(w, "failed to list deleted items", http.StatusInternalServerError)
		return
	}

	if len(result.Items) == 0 {
		result.Items = []domain.VaultItem{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  result.Items,
		"total":  result.Total,
		"limit":  result.Limit,
		"offset": result.Offset,
	})
}

func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
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

	item, err := h.vaultService.GetItem(r.Context(), userID, itemID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get item", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
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
		EncryptedData []byte   `json:"encrypted_data"` // base64 in JSON, or string
		ItemType      string   `json:"item_type"`
		Title         string   `json:"title"`
		Folder        string   `json:"folder"`
		Tags          []string `json:"tags"`
		Version       int32    `json:"version"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	input := service.UpdateVaultItemInput{
		EncryptedData: req.EncryptedData,
		ItemType:      req.ItemType,
		Title:         req.Title,
		Folder:        req.Folder,
		Tags:          req.Tags,
		Version:       req.Version,
	}

	item, err := h.vaultService.UpdateItem(r.Context(), userID, itemID, auditContextFromRequest(r), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEncryptedBlob) {
			http.Error(w, "invalid encrypted blob", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update item", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
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
		Version int32 `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	version := req.Version
	item, err := h.vaultService.DeleteItem(r.Context(), userID, itemID, auditContextFromRequest(r), version)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete item", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) RestoreItem(w http.ResponseWriter, r *http.Request) {
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
		Version int32 `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	version := req.Version

	item, err := h.vaultService.RestoreItem(r.Context(), userID, itemID, auditContextFromRequest(r), version)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to restore item", http.StatusInternalServerError)
			return
	}
	writeJSON(w, http.StatusOK, item)
}