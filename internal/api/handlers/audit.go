package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
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

func parseAuditListFilter(r *http.Request) (service.ListAuditLogsFilter, error) {
	query := r.URL.Query()
	filter := service.ListAuditLogsFilter{}

	if category := strings.TrimSpace(query.Get("category")); category != "" {
		switch category {
		case "auth", "vault", "mfa", "recovery":
			filter.ActionPrefix = category + "."
		default:
			return service.ListAuditLogsFilter{}, errors.New("invalid category")
		}
	}

	if action := strings.TrimSpace(query.Get("action")); action != "" {
		filter.Action = action
		filter.ActionPrefix = ""
	}

	if v := query.Get("since"); v != "" {
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return service.ListAuditLogsFilter{}, errors.New("invalid since")
		}
		filter.Since = &parsed
	}

	if v := query.Get("until"); v != "" {
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return service.ListAuditLogsFilter{}, errors.New("invalid until")
		}
		filter.Until = &parsed
	}

	if v := query.Get("limit"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil || parsed <= 0 {
			return service.ListAuditLogsFilter{}, errors.New("invalid limit")
		}
		filter.Limit = int32(parsed)
	}

	if v := query.Get("offset"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil || parsed < 0 {
			return service.ListAuditLogsFilter{}, errors.New("invalid offset")
		}
		filter.Offset = int32(parsed)
	}

	return filter, nil
}

func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	filter, err := parseAuditListFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logs, err := h.auditService.ListLogs(r.Context(), userID, filter)
	if err != nil {
		http.Error(w, "failed to list audit logs", http.StatusInternalServerError)
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
