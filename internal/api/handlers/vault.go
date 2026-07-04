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

	if len(req.EncryptedData) == 0 {
		http.Error(w, "encrypted_data is required", http.StatusBadRequest)
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

	item, err := h.vaultService.CreateItem(r.Context(), userID, input)
	if err != nil {
		http.Error(w, "failed to create item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	items, err := h.vaultService.ListItems(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list items", http.StatusInternalServerError)
		return
	}

	if len(items) == 0 {
		items = []domain.VaultItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(items)
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
	w.Header().Set("Content-Type", "application/json")	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}