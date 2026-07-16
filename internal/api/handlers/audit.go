package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"vault_api/internal/api/middleware"
	"vault_api/internal/requestmeta"
	"vault_api/internal/service"
)

func auditContextFromRequest(r *http.Request) service.AuditContext {
	return service.AuditContext{
		IPAddress: requestmeta.ClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
	}
}

func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	limit := int32(service.DefaultAuditLimit())
	offset := int32(0)
	if v := r.URL.Query().Get("limit"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil || parsed <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = int32(parsed)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil || parsed < 0 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		offset = int32(parsed)
	}

	logs, err := h.auditService.ListLogs(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type auditLogResponse struct {
		ID           string          `json:"id"`
		Action       string          `json:"action"`
		ResourceType string          `json:"resource_type,omitempty"`
		ResourceID   *string         `json:"resource_id,omitempty"`
		IPAddress    string          `json:"ip_address,omitempty"`
		UserAgent    string          `json:"user_agent,omitempty"`
		Metadata     json.RawMessage `json:"metadata,omitempty"`
		CreatedAt    string          `json:"created_at"`
	}

	response := make([]auditLogResponse, 0, len(logs))
	for _, entry := range logs {
		var resourceID *string
		if entry.ResourceID != nil {
			id := entry.ResourceID.String()
			resourceID = &id
		}

		response = append(response, auditLogResponse{
			ID:           entry.ID.String(),
			Action:       entry.Action,
			ResourceType: entry.ResourceType,
			ResourceID:   resourceID,
			IPAddress:    entry.IPAddress,
			UserAgent:    entry.UserAgent,
			Metadata:     entry.Metadata,
			CreatedAt:    entry.CreatedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, response)
}
